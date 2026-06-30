package usecase

import (
	"context"
	"html"
	"strings"

	"github.com/afifudin23/saepedia-api/internal/address/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
)

type AddressInput struct {
	RecipientName string
	Phone         string
	FullAddress   string
	IsPrimary     bool
}

type AddressUsecase interface {
	Create(ctx context.Context, userID string, in AddressInput) (*domain.Address, error)
	Update(ctx context.Context, userID, id string, in AddressInput) (*domain.Address, error)
	Delete(ctx context.Context, userID, id string) error
	List(ctx context.Context, userID string, limit, offset int) ([]domain.Address, int64, error)
}

type addressUsecase struct {
	repo  domain.AddressRepository
	txMgr *tx.Manager
}

func New(repo domain.AddressRepository, txMgr *tx.Manager) AddressUsecase {
	return &addressUsecase{repo: repo, txMgr: txMgr}
}

func (uc *addressUsecase) Create(ctx context.Context, userID string, in AddressInput) (*domain.Address, error) {
	a := &domain.Address{
		UserID:        userID,
		RecipientName: clean(in.RecipientName),
		Phone:         strings.TrimSpace(in.Phone),
		FullAddress:   clean(in.FullAddress),
		IsPrimary:     in.IsPrimary,
	}
	// Aturan bisnis: hanya satu alamat primary per user.
	err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
		if a.IsPrimary {
			if err := uc.repo.UnsetPrimary(ctx, userID); err != nil {
				return err
			}
		}
		return uc.repo.Create(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (uc *addressUsecase) Update(ctx context.Context, userID, id string, in AddressInput) (*domain.Address, error) {
	existing, err := uc.repo.FindForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	existing.RecipientName = clean(in.RecipientName)
	existing.Phone = strings.TrimSpace(in.Phone)
	existing.FullAddress = clean(in.FullAddress)
	existing.IsPrimary = in.IsPrimary

	err = uc.txMgr.Do(ctx, func(ctx context.Context) error {
		if existing.IsPrimary {
			if err := uc.repo.UnsetPrimary(ctx, userID); err != nil {
				return err
			}
		}
		return uc.repo.Update(ctx, existing)
	})
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *addressUsecase) Delete(ctx context.Context, userID, id string) error {
	return uc.repo.Delete(ctx, userID, id)
}

func (uc *addressUsecase) List(ctx context.Context, userID string, limit, offset int) ([]domain.Address, int64, error) {
	return uc.repo.ListByUser(ctx, userID, limit, offset)
}

func clean(s string) string { return html.EscapeString(strings.TrimSpace(s)) }
