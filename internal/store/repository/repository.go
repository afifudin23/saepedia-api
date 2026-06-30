package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/afifudin23/saepedia-api/internal/store/domain"
	"gorm.io/gorm"
)

type StoreModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      string `gorm:"type:uuid;uniqueIndex;not null"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description string `gorm:"not null;default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (StoreModel) TableName() string { return "stores" }

func (m StoreModel) toDomain() *domain.Store {
	return &domain.Store{
		ID:          m.ID,
		UserID:      m.UserID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type storeRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(ctx context.Context, s *domain.Store) error {
	model := &StoreModel{
		UserID:      s.UserID,
		Name:        s.Name,
		Description: s.Description,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return mapStoreErr(err)
	}
	s.ID = model.ID
	s.CreatedAt = model.CreatedAt
	s.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *storeRepository) Update(ctx context.Context, s *domain.Store) error {
	err := r.db.WithContext(ctx).Model(&StoreModel{}).
		Where("id = ?", s.ID).
		Updates(map[string]any{
			"name":        s.Name,
			"description": s.Description,
		}).Error
	if err != nil {
		return mapStoreErr(err)
	}
	return nil
}

func (r *storeRepository) FindByID(ctx context.Context, id string) (*domain.Store, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *storeRepository) FindByUserID(ctx context.Context, userID string) (*domain.Store, error) {
	return r.findOne(ctx, "user_id = ?", userID)
}

func (r *storeRepository) findOne(ctx context.Context, query string, arg any) (*domain.Store, error) {
	var model StoreModel
	err := r.db.WithContext(ctx).Where(query, arg).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrStoreNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *storeRepository) List(ctx context.Context, limit, offset int) ([]domain.Store, int64, error) {
	var models []StoreModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&StoreModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.Store, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out, total, nil
}

func mapStoreErr(err error) error {
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
		return err
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "name"):
		return domain.ErrStoreNameExists
	case strings.Contains(msg, "user_id"):
		return domain.ErrUserAlreadyHasStore
	default:
		return err
	}
}
