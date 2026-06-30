package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/afifudin23/saepedia-api/internal/discount/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
)

type DiscountModel struct {
	ID            string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Code          string `gorm:"uniqueIndex;not null"`
	Kind          string `gorm:"not null"`
	DiscountType  string `gorm:"not null"`
	DiscountValue int64  `gorm:"not null"`
	MaxDiscount   int64  `gorm:"not null;default:0"`
	MinSpend      int64  `gorm:"not null;default:0"`
	ExpiresAt     time.Time
	UsageLimit    *int
	UsedCount     int `gorm:"not null;default:0"`
	CreatedAt     time.Time
}

func (DiscountModel) TableName() string { return "discounts" }

func (m DiscountModel) toDomain() *domain.Discount {
	return &domain.Discount{
		ID:            m.ID,
		Code:          m.Code,
		Kind:          m.Kind,
		DiscountType:  m.DiscountType,
		DiscountValue: m.DiscountValue,
		MaxDiscount:   m.MaxDiscount,
		MinSpend:      m.MinSpend,
		ExpiresAt:     m.ExpiresAt,
		UsageLimit:    m.UsageLimit,
		UsedCount:     m.UsedCount,
		CreatedAt:     m.CreatedAt,
	}
}

type discountRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.DiscountRepository {
	return &discountRepository{db: db}
}

func (r *discountRepository) Create(ctx context.Context, d *domain.Discount) error {
	model := &DiscountModel{
		Code:          d.Code,
		Kind:          d.Kind,
		DiscountType:  d.DiscountType,
		DiscountValue: d.DiscountValue,
		MaxDiscount:   d.MaxDiscount,
		MinSpend:      d.MinSpend,
		ExpiresAt:     d.ExpiresAt,
		UsageLimit:    d.UsageLimit,
	}
	if err := tx.DB(ctx, r.db).Create(model).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return domain.ErrCodeExists
		}
		return err
	}
	d.ID = model.ID
	d.CreatedAt = model.CreatedAt
	return nil
}

func (r *discountRepository) FindByCode(ctx context.Context, code string) (*domain.Discount, error) {
	var model DiscountModel
	err := tx.DB(ctx, r.db).Where("code = ?", code).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDiscountNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *discountRepository) FindByID(ctx context.Context, id string) (*domain.Discount, error) {
	var model DiscountModel
	err := tx.DB(ctx, r.db).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrDiscountNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *discountRepository) ConsumeVoucher(ctx context.Context, id string) error {
	res := tx.DB(ctx, r.db).Exec(
		"UPDATE discounts SET used_count = used_count + 1 WHERE id = ? AND (usage_limit IS NULL OR used_count < usage_limit)",
		id,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrVoucherRejected
	}
	return nil
}

func (r *discountRepository) ListByKind(ctx context.Context, kind string, limit, offset int) ([]domain.Discount, int64, error) {
	var models []DiscountModel
	var total int64

	q := tx.DB(ctx, r.db).Model(&DiscountModel{})
	if kind != "" {
		q = q.Where("kind = ?", kind)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQ := tx.DB(ctx, r.db)
	if kind != "" {
		listQ = listQ.Where("kind = ?", kind)
	}
	if err := listQ.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	out := make([]domain.Discount, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out, total, nil
}
