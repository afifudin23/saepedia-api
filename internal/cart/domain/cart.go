package domain

import (
	"context"
	"errors"
)

var (
	ErrDifferentStore = errors.New("cart can only contain products from one store, please clear the cart first")
	ErrItemNotInCart  = errors.New("product is not in the cart")
	ErrCartEmpty      = errors.New("cart is empty")
)

// Cart adalah ringkasan keranjang yang sudah diperkaya info produk.
type Cart struct {
	ID        string
	UserID    string
	StoreID   string // "" bila cart kosong
	StoreName string
	Items     []CartItem
	Subtotal  int64
}

type CartItem struct {
	ProductID   string
	ProductName string
	Price       int64
	Quantity    int
	Stock       int
	Subtotal    int64
}

// RawItem adalah baris cart_items mentah (tanpa info produk).
type RawItem struct {
	ProductID string
	Quantity  int
}

// CartRepository mengelola tabel carts & cart_items secara mentah.
// Logika single-store dan pengayaan produk ada di usecase.
type CartRepository interface {
	GetOrCreate(ctx context.Context, userID string) (cartID string, storeID *string, err error)
	Items(ctx context.Context, cartID string) ([]RawItem, error)
	GetItemQty(ctx context.Context, cartID, productID string) (qty int, found bool, err error)
	SetStore(ctx context.Context, cartID string, storeID *string) error
	UpsertItem(ctx context.Context, cartID, productID string, qty int) error
	RemoveItem(ctx context.Context, cartID, productID string) (bool, error)
	Clear(ctx context.Context, cartID string) error
}
