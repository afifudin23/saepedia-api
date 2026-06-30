package repository

import (
	"context"
	"time"

	"github.com/afifudin23/saepedia-api/internal/review/domain"
	"gorm.io/gorm"
)

type AppReviewModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ReviewerName string `gorm:"not null"`
	Rating       int    `gorm:"not null"`
	Comment      string `gorm:"not null"`
	CreatedAt    time.Time
}

func (AppReviewModel) TableName() string { return "app_reviews" }

func (m AppReviewModel) toDomain() domain.AppReview {
	return domain.AppReview{
		ID:           m.ID,
		ReviewerName: m.ReviewerName,
		Rating:       m.Rating,
		Comment:      m.Comment,
		CreatedAt:    m.CreatedAt,
	}
}

type reviewRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, rv *domain.AppReview) error {
	model := &AppReviewModel{
		ReviewerName: rv.ReviewerName,
		Rating:       rv.Rating,
		Comment:      rv.Comment,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	rv.ID = model.ID
	rv.CreatedAt = model.CreatedAt
	return nil
}

func (r *reviewRepository) List(ctx context.Context, limit, offset int) ([]domain.AppReview, int64, error) {
	var models []AppReviewModel
	var total int64

	if err := r.db.WithContext(ctx).Model(&AppReviewModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.AppReview, 0, len(models))
	for _, m := range models {
		out = append(out, m.toDomain())
	}
	return out, total, nil
}
