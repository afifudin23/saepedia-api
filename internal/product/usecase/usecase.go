package usecase

import (
	"context"
	"html"
	"strings"

	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
)

// ProductInput dipakai untuk create/update.
type ProductInput struct {
	Name        string
	Description string
	Price       int64
	Stock       int
	ImageURL    string
}

type ProductUsecase interface {
	Create(ctx context.Context, userID string, in ProductInput) (*productdomain.Product, error)
	Update(ctx context.Context, userID, productID string, in ProductInput) (*productdomain.Product, error)
	Delete(ctx context.Context, userID, productID string) error
	ListMine(ctx context.Context, userID string, limit, offset int) ([]productdomain.Product, int64, error)
	ListPublic(ctx context.Context, search string, limit, offset int) ([]productdomain.Product, int64, error)
	GetPublic(ctx context.Context, id string) (*productdomain.Product, error)
}

type productUsecase struct {
	repo      productdomain.ProductRepository
	storeRepo storedomain.StoreRepository
}

func New(repo productdomain.ProductRepository, storeRepo storedomain.StoreRepository) ProductUsecase {
	return &productUsecase{repo: repo, storeRepo: storeRepo}
}

func (uc *productUsecase) Create(ctx context.Context, userID string, in ProductInput) (*productdomain.Product, error) {
	// Seller hanya boleh membuat produk di bawah tokonya sendiri.
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err // ErrStoreNotFound → seller belum punya toko
	}

	product := &productdomain.Product{
		StoreID:     store.ID,
		Name:        strings.TrimSpace(in.Name),
		Description: html.EscapeString(strings.TrimSpace(in.Description)),
		Price:       in.Price,
		Stock:       in.Stock,
		ImageURL:    strings.TrimSpace(in.ImageURL),
	}
	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (uc *productUsecase) Update(ctx context.Context, userID, productID string, in ProductInput) (*productdomain.Product, error) {
	product, err := uc.ownedProduct(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	product.Name = strings.TrimSpace(in.Name)
	product.Description = html.EscapeString(strings.TrimSpace(in.Description))
	product.Price = in.Price
	product.Stock = in.Stock
	product.ImageURL = strings.TrimSpace(in.ImageURL)
	if err := uc.repo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (uc *productUsecase) Delete(ctx context.Context, userID, productID string) error {
	product, err := uc.ownedProduct(ctx, userID, productID)
	if err != nil {
		return err
	}
	return uc.repo.Delete(ctx, product.ID)
}

func (uc *productUsecase) ListMine(ctx context.Context, userID string, limit, offset int) ([]productdomain.Product, int64, error) {
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return uc.repo.ListByStore(ctx, store.ID, limit, offset)
}

func (uc *productUsecase) ListPublic(ctx context.Context, search string, limit, offset int) ([]productdomain.Product, int64, error) {
	return uc.repo.ListPublic(ctx, strings.TrimSpace(search), limit, offset)
}

func (uc *productUsecase) GetPublic(ctx context.Context, id string) (*productdomain.Product, error) {
	return uc.repo.FindByID(ctx, id)
}

// ownedProduct memastikan produk ada DAN milik toko seller yang sedang login.
func (uc *productUsecase) ownedProduct(ctx context.Context, userID, productID string) (*productdomain.Product, error) {
	product, err := uc.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if product.StoreID != store.ID {
		return nil, productdomain.ErrNotProductOwner
	}
	return product, nil
}
