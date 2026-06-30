package repository

import (
	"context"
	"time"

	"github.com/afifudin23/saepedia-api/internal/cart/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartModel struct {
	ID        string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string  `gorm:"type:uuid;uniqueIndex;not null"`
	StoreID   *string `gorm:"type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CartModel) TableName() string { return "carts" }

type CartItemModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CartID    string `gorm:"type:uuid;not null;index"`
	ProductID string `gorm:"type:uuid;not null"`
	Quantity  int    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CartItemModel) TableName() string { return "cart_items" }

type cartRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreate(ctx context.Context, userID string) (string, *string, error) {
	model := CartModel{UserID: userID}
	if err := tx.DB(ctx, r.db).Where(CartModel{UserID: userID}).FirstOrCreate(&model).Error; err != nil {
		return "", nil, err
	}
	return model.ID, model.StoreID, nil
}

func (r *cartRepository) Items(ctx context.Context, cartID string) ([]domain.RawItem, error) {
	var models []CartItemModel
	err := tx.DB(ctx, r.db).Where("cart_id = ?", cartID).Order("created_at ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.RawItem, 0, len(models))
	for _, m := range models {
		out = append(out, domain.RawItem{ProductID: m.ProductID, Quantity: m.Quantity})
	}
	return out, nil
}

func (r *cartRepository) GetItemQty(ctx context.Context, cartID, productID string) (int, bool, error) {
	var model CartItemModel
	err := tx.DB(ctx, r.db).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return model.Quantity, true, nil
}

func (r *cartRepository) SetStore(ctx context.Context, cartID string, storeID *string) error {
	return tx.DB(ctx, r.db).Model(&CartModel{}).Where("id = ?", cartID).
		Update("store_id", storeID).Error
}

func (r *cartRepository) UpsertItem(ctx context.Context, cartID, productID string, qty int) error {
	item := CartItemModel{CartID: cartID, ProductID: productID, Quantity: qty}
	return tx.DB(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "cart_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]any{"quantity": qty, "updated_at": time.Now()}),
	}).Create(&item).Error
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID, productID string) (bool, error) {
	res := tx.DB(ctx, r.db).Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&CartItemModel{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *cartRepository) Clear(ctx context.Context, cartID string) error {
	return tx.DB(ctx, r.db).Where("cart_id = ?", cartID).Delete(&CartItemModel{}).Error
}
