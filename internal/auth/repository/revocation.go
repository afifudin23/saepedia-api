package repository

import (
	"context"
	"time"

	"github.com/afifudin23/saepedia-api/internal/auth/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RevokedTokenModel struct {
	JTI       string `gorm:"primaryKey"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (RevokedTokenModel) TableName() string { return "revoked_tokens" }

type revocationRepository struct {
	db *gorm.DB
}

func NewRevocation(db *gorm.DB) domain.TokenRevoker {
	return &revocationRepository{db: db}
}

func (r *revocationRepository) Revoke(ctx context.Context, jti string, exp time.Time) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&RevokedTokenModel{JTI: jti, ExpiresAt: exp}).Error
}

func (r *revocationRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RevokedTokenModel{}).Where("jti = ?", jti).Count(&count).Error
	return count > 0, err
}
