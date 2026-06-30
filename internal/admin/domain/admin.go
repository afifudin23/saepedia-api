package domain

import (
	"context"
	"time"
)

// Summary adalah ringkasan monitoring marketplace untuk dashboard admin.
type Summary struct {
	Users            int64
	Stores           int64
	Products         int64
	Orders           int64
	OrdersByStatus   map[string]int64
	Vouchers         int64
	Promos           int64
	AvailableJobs    int64 // order status Menunggu Pengirim
	ActiveDeliveries int64 // order status Sedang Dikirim
	OverdueOrders    int64 // diisi usecase (butuh SLA per metode)
}

type UserRow struct {
	ID        string
	Email     string
	IsAdmin   bool
	Roles     string
	CreatedAt time.Time
}

type StoreRow struct {
	ID           string
	Name         string
	Owner        string
	ProductCount int64
	CreatedAt    time.Time
}

type ProductRow struct {
	ID        string
	Name      string
	StoreName string
	Price     int64
	Stock     int
	CreatedAt time.Time
}

type OrderRow struct {
	ID             string
	BuyerEmail     string
	StoreName      string
	Status         string
	DeliveryMethod string
	Total          int64
	CreatedAt      time.Time
}

type DeliveryRow struct {
	OrderID     string
	StoreName   string
	DriverEmail string
	Status      string
	DeliveryFee int64
	Earning     int64
}

type MonitorRepository interface {
	Summary(ctx context.Context) (*Summary, error)
	ListUsers(ctx context.Context, limit, offset int) ([]UserRow, int64, error)
	ListStores(ctx context.Context, limit, offset int) ([]StoreRow, int64, error)
	ListProducts(ctx context.Context, limit, offset int) ([]ProductRow, int64, error)
	ListOrders(ctx context.Context, limit, offset int) ([]OrderRow, int64, error)
	ListDeliveries(ctx context.Context, limit, offset int) ([]DeliveryRow, int64, error)
}
