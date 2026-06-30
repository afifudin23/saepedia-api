package usecase

import (
	"context"
	"errors"
	"html"
	"strings"

	"github.com/afifudin23/saepedia-api/internal/store/domain"
)

type StoreUsecase interface {
	// CreateOrUpdate membuat toko bila belum ada, atau memperbarui toko milik seller.
	CreateOrUpdate(ctx context.Context, userID, name, description string) (*domain.Store, error)
	GetMine(ctx context.Context, userID string) (*domain.Store, error)
	GetByID(ctx context.Context, id string) (*domain.Store, error)
	List(ctx context.Context, limit, offset int) ([]domain.Store, int64, error)
}

type storeUsecase struct {
	repo domain.StoreRepository
}

func New(repo domain.StoreRepository) StoreUsecase {
	return &storeUsecase{repo: repo}
}

func (uc *storeUsecase) CreateOrUpdate(ctx context.Context, userID, name, description string) (*domain.Store, error) {
	name = strings.TrimSpace(name)
	description = html.EscapeString(strings.TrimSpace(description))

	existing, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrStoreNotFound) {
		return nil, err
	}

	// Update toko yang sudah ada (seller hanya boleh mengelola toko sendiri).
	if existing != nil {
		existing.Name = name
		existing.Description = description
		if err := uc.repo.Update(ctx, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	store := &domain.Store{
		UserID:      userID,
		Name:        name,
		Description: description,
	}
	if err := uc.repo.Create(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

func (uc *storeUsecase) GetMine(ctx context.Context, userID string) (*domain.Store, error) {
	return uc.repo.FindByUserID(ctx, userID)
}

func (uc *storeUsecase) GetByID(ctx context.Context, id string) (*domain.Store, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *storeUsecase) List(ctx context.Context, limit, offset int) ([]domain.Store, int64, error) {
	return uc.repo.List(ctx, limit, offset)
}
