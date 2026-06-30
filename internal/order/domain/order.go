package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCartEmpty         = errors.New("cart is empty")
	ErrInvalidDelivery   = errors.New("invalid delivery method")
	ErrInsufficientStock = errors.New("insufficient product stock")
	ErrInsufficientFunds = errors.New("insufficient wallet balance")
	ErrAddressNotFound   = errors.New("delivery address not found")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidTransition = errors.New("order status cannot be changed from its current state")
	ErrDiscountRejected  = errors.New("discount code can no longer be used")
)

// Status utama order (wajib tampil di UI).
const (
	StatusDikemas       = "Sedang Dikemas"
	StatusMenungguKirim = "Menunggu Pengirim"
	StatusDikirim       = "Sedang Dikirim"
	StatusSelesai       = "Pesanan Selesai"
	StatusDikembalikan  = "Dikembalikan"
)

// Metode pengiriman & tarifnya (rupiah). Berbeda per metode.
const (
	DeliveryInstant = "instant"
	DeliveryNextDay = "next_day"
	DeliveryRegular = "regular"
)

// TaxRatePercent — PPN 12% dihitung dari (subtotal - discount).
const TaxRatePercent = 12

// DriverEarningPercent — driver mendapat 80% dari ongkir untuk job yang selesai.
const DriverEarningPercent = 80

// DeliveryFee mengembalikan tarif kirim untuk metode tertentu.
func DeliveryFee(method string) (int64, bool) {
	switch method {
	case DeliveryInstant:
		return 20000, true
	case DeliveryNextDay:
		return 10000, true
	case DeliveryRegular:
		return 5000, true
	default:
		return 0, false
	}
}

// SLA mengembalikan batas waktu penyelesaian order per metode pengiriman.
// Order yang melewati (created_at + SLA) tanpa selesai dianggap overdue.
func SLA(method string) time.Duration {
	switch method {
	case DeliveryInstant:
		return 1 * 24 * time.Hour
	case DeliveryNextDay:
		return 2 * 24 * time.Hour
	case DeliveryRegular:
		return 3 * 24 * time.Hour
	default:
		return 3 * 24 * time.Hour
	}
}

// CalcTax menghitung PPN 12% dari basis (subtotal - discount), floor ke rupiah.
func CalcTax(base int64) int64 {
	if base < 0 {
		base = 0
	}
	return base * TaxRatePercent / 100
}

// CalcDriverEarning menghitung pendapatan driver dari ongkir.
func CalcDriverEarning(deliveryFee int64) int64 {
	return deliveryFee * DriverEarningPercent / 100
}

type Order struct {
	ID             string
	BuyerID        string
	StoreID        string
	StoreName      string
	BuyerEmail     string
	RecipientName  string
	Phone          string
	FullAddress    string
	DeliveryMethod string
	Subtotal       int64
	Discount       int64
	DiscountCode   string
	DeliveryFee    int64
	Tax            int64
	Total          int64
	Status         string
	DriverID       string
	DriverEarning  int64
	Refunded       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Items         []OrderItem
	StatusHistory []OrderStatus
}

type OrderItem struct {
	ProductID   string
	ProductName string
	Price       int64
	Quantity    int
	Subtotal    int64
}

type OrderStatus struct {
	Status    string
	Note      string
	CreatedAt time.Time
}

// BuyerReport — ringkasan pengeluaran buyer.
type BuyerReport struct {
	TotalOrders   int
	TotalSpent    int64 // total order yang TIDAK dikembalikan
	TotalRefunded int64 // total order yang dikembalikan
	CountByStatus map[string]int
}

// SellerReport — ringkasan pendapatan seller.
// Pendapatan diakui saat order "Pesanan Selesai"; order dikembalikan tidak dihitung.
type SellerReport struct {
	TotalOrders     int
	CompletedOrders int
	TotalRevenue    int64 // Σ (subtotal - discount) order selesai
	TotalRefunded   int64 // Σ total order dikembalikan
	CountByStatus   map[string]int
}

// OrderRepository hanya menyediakan operasi data. Semua keputusan bisnis
// (validasi, perhitungan, urutan langkah, aturan transisi status) ada di usecase
// dan dibungkus satu transaksi via tx.Manager.
type OrderRepository interface {
	// Create menyimpan order beserta item-nya (status diisi usecase).
	Create(ctx context.Context, o *Order) error
	AddStatusHistory(ctx context.Context, orderID, status, note string) error
	// UpdateStatusGuarded mengubah status secara atomik hanya bila status saat ini = from.
	UpdateStatusGuarded(ctx context.Context, orderID, from, to string) (bool, error)
	// MarkRefunded menandai order Dikembalikan + refunded_at (anti double refund).
	MarkRefunded(ctx context.Context, orderID string) error
	// GetForUpdate mengunci baris order (dipakai di dalam transaksi refund).
	GetForUpdate(ctx context.Context, orderID string) (*Order, error)
	Items(ctx context.Context, orderID string) ([]OrderItem, error)

	FindByIDForBuyer(ctx context.Context, buyerID, orderID string) (*Order, error)
	FindByIDForStore(ctx context.Context, storeID, orderID string) (*Order, error)
	ListByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]Order, int64, error)
	ListByStore(ctx context.Context, storeID string, limit, offset int) ([]Order, int64, error)

	BuyerReport(ctx context.Context, buyerID string) (*BuyerReport, error)
	SellerReport(ctx context.Context, storeID string) (*SellerReport, error)

	// ListRefundCandidates mengembalikan order yang belum final & belum di-refund.
	// Penyaringan SLA/overdue dilakukan di usecase (pakai clock virtual).
	ListRefundCandidates(ctx context.Context) ([]Order, error)
}
