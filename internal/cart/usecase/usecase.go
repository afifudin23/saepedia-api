package usecase

import (
	"context"

	cartdomain "github.com/afifudin23/saepedia-api/internal/cart/domain"
	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
)

type CartUsecase interface {
	Get(ctx context.Context, userID string) (*cartdomain.Cart, error)
	AddItem(ctx context.Context, userID, productID string, qty int) (*cartdomain.Cart, error)
	UpdateItem(ctx context.Context, userID, productID string, qty int) (*cartdomain.Cart, error)
	RemoveItem(ctx context.Context, userID, productID string) (*cartdomain.Cart, error)
	Clear(ctx context.Context, userID string) error
}

type cartUsecase struct {
	repo        cartdomain.CartRepository
	productRepo productdomain.ProductRepository
}

func New(repo cartdomain.CartRepository, productRepo productdomain.ProductRepository) CartUsecase {
	return &cartUsecase{repo: repo, productRepo: productRepo}
}

func (uc *cartUsecase) AddItem(ctx context.Context, userID, productID string, qty int) (*cartdomain.Cart, error) {
	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	cartID, storeID, err := uc.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Single-store rule: cart hanya boleh berisi produk dari satu toko.
	if storeID != nil && *storeID != product.StoreID {
		return nil, cartdomain.ErrDifferentStore
	}

	currentQty, _, err := uc.repo.GetItemQty(ctx, cartID, productID)
	if err != nil {
		return nil, err
	}
	newQty := currentQty + qty
	if newQty > product.Stock {
		return nil, productdomain.ErrInsufficientStock
	}

	if storeID == nil {
		sid := product.StoreID
		if err := uc.repo.SetStore(ctx, cartID, &sid); err != nil {
			return nil, err
		}
	}
	if err := uc.repo.UpsertItem(ctx, cartID, productID, newQty); err != nil {
		return nil, err
	}

	return uc.buildCart(ctx, userID)
}

func (uc *cartUsecase) UpdateItem(ctx context.Context, userID, productID string, qty int) (*cartdomain.Cart, error) {
	cartID, _, err := uc.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	_, found, err := uc.repo.GetItemQty(ctx, cartID, productID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, cartdomain.ErrItemNotInCart
	}

	product, err := uc.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if qty > product.Stock {
		return nil, productdomain.ErrInsufficientStock
	}

	if err := uc.repo.UpsertItem(ctx, cartID, productID, qty); err != nil {
		return nil, err
	}
	return uc.buildCart(ctx, userID)
}

func (uc *cartUsecase) RemoveItem(ctx context.Context, userID, productID string) (*cartdomain.Cart, error) {
	cartID, _, err := uc.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	removed, err := uc.repo.RemoveItem(ctx, cartID, productID)
	if err != nil {
		return nil, err
	}
	if !removed {
		return nil, cartdomain.ErrItemNotInCart
	}

	// Bila cart jadi kosong, reset store_id supaya bisa diisi toko lain.
	items, err := uc.repo.Items(ctx, cartID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		if err := uc.repo.SetStore(ctx, cartID, nil); err != nil {
			return nil, err
		}
	}

	return uc.buildCart(ctx, userID)
}

func (uc *cartUsecase) Clear(ctx context.Context, userID string) error {
	cartID, _, err := uc.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	if err := uc.repo.Clear(ctx, cartID); err != nil {
		return err
	}
	return uc.repo.SetStore(ctx, cartID, nil)
}

func (uc *cartUsecase) Get(ctx context.Context, userID string) (*cartdomain.Cart, error) {
	return uc.buildCart(ctx, userID)
}

// buildCart membaca item mentah lalu memperkaya dengan info produk terkini.
func (uc *cartUsecase) buildCart(ctx context.Context, userID string) (*cartdomain.Cart, error) {
	cartID, storeID, err := uc.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}
	rawItems, err := uc.repo.Items(ctx, cartID)
	if err != nil {
		return nil, err
	}

	cart := &cartdomain.Cart{ID: cartID, UserID: userID, Items: []cartdomain.CartItem{}}
	if storeID != nil {
		cart.StoreID = *storeID
	}

	for _, raw := range rawItems {
		product, err := uc.productRepo.FindByID(ctx, raw.ProductID)
		if err != nil {
			// Produk sudah dihapus seller — lewati dari ringkasan.
			continue
		}
		sub := product.Price * int64(raw.Quantity)
		cart.Items = append(cart.Items, cartdomain.CartItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Price:       product.Price,
			Quantity:    raw.Quantity,
			Stock:       product.Stock,
			Subtotal:    sub,
		})
		cart.Subtotal += sub
		if cart.StoreName == "" {
			cart.StoreName = product.StoreName
		}
	}

	return cart, nil
}
