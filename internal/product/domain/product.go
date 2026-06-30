package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrNotProductOwner   = errors.New("you don't own this product")
	ErrInsufficientStock = errors.New("insufficient product stock")
)

type Product struct {
	ID          string
	StoreID     string
	Name        string
	Description string
	Price       int64
	Stock       int
	ImageURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// StoreName diisi saat query publik (read-only, hasil join).
	StoreName string
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Product, error)
	ListByStore(ctx context.Context, storeID string, limit, offset int) ([]Product, int64, error)
	ListPublic(ctx context.Context, search string, limit, offset int) ([]Product, int64, error)

	// DecrementStock mengurangi stok secara aman (gagal bila stok kurang).
	// Operasi atomik tingkat-data; dipanggil dari dalam transaksi checkout.
	DecrementStock(ctx context.Context, productID string, qty int) error
	// RestoreStock menambah kembali stok (dipakai saat refund/return).
	RestoreStock(ctx context.Context, productID string, qty int) error
}
