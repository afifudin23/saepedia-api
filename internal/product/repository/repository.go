package repository

import (
	"context"
	"errors"
	"time"

	"github.com/afifudin23/saepedia-api/internal/product/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
)

type ProductModel struct {
	ID          string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	StoreID     string `gorm:"type:uuid;not null;index"`
	Name        string `gorm:"not null"`
	Description string `gorm:"not null;default:''"`
	Price       int64  `gorm:"not null"`
	Stock       int    `gorm:"not null;default:0"`
	ImageURL    string `gorm:"not null;default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (ProductModel) TableName() string { return "products" }

// productRow dipakai untuk hasil query yang membawa nama toko (join).
type productRow struct {
	ProductModel
	StoreName string
}

func (m ProductModel) toDomain() *domain.Product {
	return &domain.Product{
		ID:          m.ID,
		StoreID:     m.StoreID,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Stock:       m.Stock,
		ImageURL:    m.ImageURL,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (r productRow) toDomain() *domain.Product {
	p := r.ProductModel.toDomain()
	p.StoreName = r.StoreName
	return p
}

type productRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	model := &ProductModel{
		StoreID:     p.StoreID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
	}
	if err := tx.DB(ctx, r.db).Create(model).Error; err != nil {
		return err
	}
	p.ID = model.ID
	p.CreatedAt = model.CreatedAt
	p.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	return tx.DB(ctx, r.db).Model(&ProductModel{}).
		Where("id = ?", p.ID).
		Updates(map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"price":       p.Price,
			"stock":       p.Stock,
			"image_url":   p.ImageURL,
		}).Error
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	return tx.DB(ctx, r.db).Where("id = ?", id).Delete(&ProductModel{}).Error
}

func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var row productRow
	err := tx.DB(ctx, r.db).
		Table("products").
		Select("products.*, stores.name AS store_name").
		Joins("JOIN stores ON stores.id = products.store_id").
		Where("products.id = ?", id).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *productRepository) ListByStore(ctx context.Context, storeID string, limit, offset int) ([]domain.Product, int64, error) {
	var models []ProductModel
	var total int64

	if err := tx.DB(ctx, r.db).Model(&ProductModel{}).Where("store_id = ?", storeID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.DB(ctx, r.db).
		Where("store_id = ?", storeID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.Product, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out, total, nil
}

func (r *productRepository) DecrementStock(ctx context.Context, productID string, qty int) error {
	res := tx.DB(ctx, r.db).Exec(
		"UPDATE products SET stock = stock - ?, updated_at = NOW() WHERE id = ? AND stock >= ?",
		qty, productID, qty,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrInsufficientStock
	}
	return nil
}

func (r *productRepository) RestoreStock(ctx context.Context, productID string, qty int) error {
	return tx.DB(ctx, r.db).Exec(
		"UPDATE products SET stock = stock + ?, updated_at = NOW() WHERE id = ?",
		qty, productID,
	).Error
}

func (r *productRepository) ListPublic(ctx context.Context, search string, limit, offset int) ([]domain.Product, int64, error) {
	base := tx.DB(ctx, r.db).
		Table("products").
		Joins("JOIN stores ON stores.id = products.store_id")

	if search != "" {
		// ILIKE dengan parameter — aman dari SQL injection (parameterized).
		base = base.Where("products.name ILIKE ?", "%"+search+"%")
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []productRow
	err := base.
		Select("products.*, stores.name AS store_name").
		Order("products.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		out = append(out, *row.toDomain())
	}
	return out, total, nil
}
