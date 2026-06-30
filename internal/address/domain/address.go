package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAddressNotFound = errors.New("address not found")
)

type Address struct {
	ID            string
	UserID        string
	RecipientName string
	Phone         string
	FullAddress   string
	IsPrimary     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AddressRepository interface {
	Create(ctx context.Context, a *Address) error
	Update(ctx context.Context, a *Address) error
	Delete(ctx context.Context, userID, id string) error
	// UnsetPrimary menonaktifkan flag primary semua alamat milik user.
	UnsetPrimary(ctx context.Context, userID string) error
	// FindForUser mengembalikan alamat HANYA bila milik user tsb.
	FindForUser(ctx context.Context, userID, id string) (*Address, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Address, int64, error)
}
