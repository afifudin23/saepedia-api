package domain

import (
	"context"
	"errors"
	"time"

	"github.com/afifudin23/saepedia-api/pkg/clock"
)

var (
	ErrDiscountNotFound = errors.New("discount code not found")
	ErrDiscountExpired  = errors.New("discount code has expired")
	ErrDiscountUsedUp   = errors.New("voucher has no remaining usage")
	ErrMinSpendNotMet   = errors.New("subtotal does not meet the minimum spend for this discount")
	ErrCodeExists       = errors.New("discount code already exists")
	ErrVoucherRejected  = errors.New("voucher can no longer be used")
)

const (
	KindVoucher = "voucher"
	KindPromo   = "promo"

	TypePercent = "percent"
	TypeFixed   = "fixed"
)

// Discount mewakili Voucher (punya kuota pemakaian) atau Promo (tanpa kuota).
type Discount struct {
	ID            string
	Code          string
	Kind          string
	DiscountType  string
	DiscountValue int64
	MaxDiscount   int64 // cap untuk percent; 0 = tanpa cap
	MinSpend      int64
	ExpiresAt     time.Time
	UsageLimit    *int // hanya voucher
	UsedCount     int
	CreatedAt     time.Time
}

// CalcAmount menghitung nominal potongan dari subtotal (tidak melebihi subtotal).
func (d *Discount) CalcAmount(subtotal int64) int64 {
	var amount int64
	switch d.DiscountType {
	case TypePercent:
		amount = subtotal * d.DiscountValue / 100
		if d.MaxDiscount > 0 && amount > d.MaxDiscount {
			amount = d.MaxDiscount
		}
	case TypeFixed:
		amount = d.DiscountValue
	}
	if amount > subtotal {
		amount = subtotal
	}
	if amount < 0 {
		amount = 0
	}
	return amount
}

// Validate mengecek kelayakan diskon terhadap subtotal & waktu virtual saat ini.
func (d *Discount) Validate(subtotal int64) error {
	if clock.Now().After(d.ExpiresAt) {
		return ErrDiscountExpired
	}
	if d.Kind == KindVoucher && d.UsageLimit != nil && d.UsedCount >= *d.UsageLimit {
		return ErrDiscountUsedUp
	}
	if subtotal < d.MinSpend {
		return ErrMinSpendNotMet
	}
	return nil
}

type DiscountRepository interface {
	Create(ctx context.Context, d *Discount) error
	FindByCode(ctx context.Context, code string) (*Discount, error)
	FindByID(ctx context.Context, id string) (*Discount, error)
	ListByKind(ctx context.Context, kind string, limit, offset int) ([]Discount, int64, error)
	// ConsumeVoucher menaikkan used_count secara atomik (gagal bila kuota habis).
	// Dipanggil dari dalam transaksi checkout.
	ConsumeVoucher(ctx context.Context, id string) error
}
