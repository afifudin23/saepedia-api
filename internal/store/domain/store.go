package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrStoreNotFound       = errors.New("store not found")
	ErrStoreNameExists     = errors.New("store name already exists")
	ErrUserAlreadyHasStore = errors.New("seller already has a store")
)

type Store struct {
	ID          string
	UserID      string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type StoreRepository interface {
	Create(ctx context.Context, s *Store) error
	Update(ctx context.Context, s *Store) error
	FindByID(ctx context.Context, id string) (*Store, error)
	FindByUserID(ctx context.Context, userID string) (*Store, error)
	List(ctx context.Context, limit, offset int) ([]Store, int64, error)
}
