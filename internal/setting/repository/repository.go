// Package repository menyimpan setting global (key-value), mis. offset simulasi waktu.
package repository

import (
	"context"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const keyTimeOffsetSeconds = "time_offset_seconds"

type SettingModel struct {
	Key       string `gorm:"primaryKey"`
	Value     string `gorm:"not null"`
	UpdatedAt time.Time
}

func (SettingModel) TableName() string { return "app_settings" }

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetTimeOffset membaca offset simulasi waktu (default 0 bila belum ada).
func (r *Repository) GetTimeOffset(ctx context.Context) (time.Duration, error) {
	var model SettingModel
	err := r.db.WithContext(ctx).Where("key = ?", keyTimeOffsetSeconds).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	secs, _ := strconv.ParseInt(model.Value, 10, 64)
	return time.Duration(secs) * time.Second, nil
}

// SetTimeOffset menyimpan offset simulasi waktu.
func (r *Repository) SetTimeOffset(ctx context.Context, d time.Duration) error {
	secs := int64(d / time.Second)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{"value": strconv.FormatInt(secs, 10), "updated_at": time.Now()}),
	}).Create(&SettingModel{Key: keyTimeOffsetSeconds, Value: strconv.FormatInt(secs, 10)}).Error
}
