package repository

import (
	"context"
	"errors"
	"time"

	"github.com/afifudin23/saepedia-api/internal/address/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
)

type AddressModel struct {
	ID            string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID        string `gorm:"type:uuid;not null;index"`
	RecipientName string `gorm:"not null"`
	Phone         string `gorm:"not null"`
	FullAddress   string `gorm:"not null"`
	IsPrimary     bool   `gorm:"not null;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (AddressModel) TableName() string { return "addresses" }

func (m AddressModel) toDomain() *domain.Address {
	return &domain.Address{
		ID: m.ID, UserID: m.UserID, RecipientName: m.RecipientName,
		Phone: m.Phone, FullAddress: m.FullAddress, IsPrimary: m.IsPrimary,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

type addressRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(ctx context.Context, a *domain.Address) error {
	model := &AddressModel{
		UserID: a.UserID, RecipientName: a.RecipientName,
		Phone: a.Phone, FullAddress: a.FullAddress, IsPrimary: a.IsPrimary,
	}
	if err := tx.DB(ctx, r.db).Create(model).Error; err != nil {
		return err
	}
	a.ID = model.ID
	a.CreatedAt = model.CreatedAt
	a.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *addressRepository) Update(ctx context.Context, a *domain.Address) error {
	return tx.DB(ctx, r.db).Model(&AddressModel{}).
		Where("id = ? AND user_id = ?", a.ID, a.UserID).
		Updates(map[string]any{
			"recipient_name": a.RecipientName,
			"phone":          a.Phone,
			"full_address":   a.FullAddress,
			"is_primary":     a.IsPrimary,
		}).Error
}

func (r *addressRepository) UnsetPrimary(ctx context.Context, userID string) error {
	return tx.DB(ctx, r.db).Model(&AddressModel{}).
		Where("user_id = ?", userID).Update("is_primary", false).Error
}

func (r *addressRepository) Delete(ctx context.Context, userID, id string) error {
	res := tx.DB(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).Delete(&AddressModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func (r *addressRepository) FindForUser(ctx context.Context, userID, id string) (*domain.Address, error) {
	var model AddressModel
	err := tx.DB(ctx, r.db).Where("id = ? AND user_id = ?", id, userID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAddressNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *addressRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Address, int64, error) {
	var models []AddressModel
	var total int64

	if err := tx.DB(ctx, r.db).Model(&AddressModel{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.DB(ctx, r.db).
		Where("user_id = ?", userID).
		Order("is_primary DESC, created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.Address, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out, total, nil
}
