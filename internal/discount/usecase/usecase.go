package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/afifudin23/saepedia-api/internal/discount/domain"
)

// GenerateInput dipakai admin untuk membuat voucher/promo.
type GenerateInput struct {
	Code          string
	DiscountType  string
	DiscountValue int64
	MaxDiscount   int64
	MinSpend      int64
	ExpiresAt     time.Time
	UsageLimit    *int // hanya relevan untuk voucher
}

// ValidationResult adalah hasil validasi diskon untuk checkout.
type ValidationResult struct {
	ID     string
	Code   string
	Kind   string
	Amount int64
}

type DiscountUsecase interface {
	Generate(ctx context.Context, kind string, in GenerateInput) (*domain.Discount, error)
	GetByID(ctx context.Context, id string) (*domain.Discount, error)
	List(ctx context.Context, kind string, limit, offset int) ([]domain.Discount, int64, error)
	// Validate memvalidasi kode terhadap subtotal & menghitung potongan.
	Validate(ctx context.Context, code string, subtotal int64) (*ValidationResult, error)
	// ConsumeVoucher memakai satu kuota voucher (dipanggil di dalam transaksi checkout).
	ConsumeVoucher(ctx context.Context, id string) error
}

type discountUsecase struct {
	repo domain.DiscountRepository
}

func New(repo domain.DiscountRepository) DiscountUsecase {
	return &discountUsecase{repo: repo}
}

func (uc *discountUsecase) Generate(ctx context.Context, kind string, in GenerateInput) (*domain.Discount, error) {
	d := &domain.Discount{
		Code:          strings.ToUpper(strings.TrimSpace(in.Code)),
		Kind:          kind,
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		MaxDiscount:   in.MaxDiscount,
		MinSpend:      in.MinSpend,
		ExpiresAt:     in.ExpiresAt,
	}
	// Promo tidak punya kuota pemakaian.
	if kind == domain.KindVoucher {
		d.UsageLimit = in.UsageLimit
	}
	if err := uc.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (uc *discountUsecase) GetByID(ctx context.Context, id string) (*domain.Discount, error) {
	return uc.repo.FindByID(ctx, id)
}

func (uc *discountUsecase) List(ctx context.Context, kind string, limit, offset int) ([]domain.Discount, int64, error) {
	return uc.repo.ListByKind(ctx, kind, limit, offset)
}

func (uc *discountUsecase) ConsumeVoucher(ctx context.Context, id string) error {
	return uc.repo.ConsumeVoucher(ctx, id)
}

func (uc *discountUsecase) Validate(ctx context.Context, code string, subtotal int64) (*ValidationResult, error) {
	d, err := uc.repo.FindByCode(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err != nil {
		return nil, err
	}
	if err := d.Validate(subtotal); err != nil {
		return nil, err
	}
	return &ValidationResult{
		ID:     d.ID,
		Code:   d.Code,
		Kind:   d.Kind,
		Amount: d.CalcAmount(subtotal),
	}, nil
}
