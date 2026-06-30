package repository

import (
	"context"

	"github.com/afifudin23/saepedia-api/internal/admin/domain"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	"gorm.io/gorm"
)

type monitorRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.MonitorRepository {
	return &monitorRepository{db: db}
}

func (r *monitorRepository) Summary(ctx context.Context) (*domain.Summary, error) {
	db := r.db.WithContext(ctx)
	s := &domain.Summary{OrdersByStatus: map[string]int64{}}

	count := func(table string) (int64, error) {
		var n int64
		err := db.Table(table).Count(&n).Error
		return n, err
	}

	var err error
	if s.Users, err = count("users"); err != nil {
		return nil, err
	}
	if s.Stores, err = count("stores"); err != nil {
		return nil, err
	}
	if s.Products, err = count("products"); err != nil {
		return nil, err
	}
	if s.Orders, err = count("orders"); err != nil {
		return nil, err
	}

	if err = db.Table("discounts").Where("kind = ?", "voucher").Count(&s.Vouchers).Error; err != nil {
		return nil, err
	}
	if err = db.Table("discounts").Where("kind = ?", "promo").Count(&s.Promos).Error; err != nil {
		return nil, err
	}
	if err = db.Table("orders").Where("status = ?", orderdomain.StatusMenungguKirim).Count(&s.AvailableJobs).Error; err != nil {
		return nil, err
	}
	if err = db.Table("orders").Where("status = ?", orderdomain.StatusDikirim).Count(&s.ActiveDeliveries).Error; err != nil {
		return nil, err
	}

	var statusRows []struct {
		Status string
		Count  int64
	}
	if err = db.Table("orders").Select("status, COUNT(*) AS count").Group("status").Scan(&statusRows).Error; err != nil {
		return nil, err
	}
	for _, row := range statusRows {
		s.OrdersByStatus[row.Status] = row.Count
	}

	return s, nil
}

func (r *monitorRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.UserRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("users").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.UserRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.email, u.is_admin, u.created_at,
			COALESCE(string_agg(ur.role, ',' ORDER BY ur.role), '') AS roles
		FROM users u
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?`, limit, offset).Scan(&rows).Error
	return rows, total, err
}

func (r *monitorRepository) ListStores(ctx context.Context, limit, offset int) ([]domain.StoreRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("stores").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.StoreRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT s.id, s.name, u.email AS owner, s.created_at,
			(SELECT COUNT(*) FROM products p WHERE p.store_id = s.id) AS product_count
		FROM stores s
		JOIN users u ON u.id = s.user_id
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?`, limit, offset).Scan(&rows).Error
	return rows, total, err
}

func (r *monitorRepository) ListProducts(ctx context.Context, limit, offset int) ([]domain.ProductRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("products").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.ProductRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id, p.name, st.name AS store_name, p.price, p.stock, p.created_at
		FROM products p
		JOIN stores st ON st.id = p.store_id
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`, limit, offset).Scan(&rows).Error
	return rows, total, err
}

func (r *monitorRepository) ListOrders(ctx context.Context, limit, offset int) ([]domain.OrderRow, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Table("orders").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.OrderRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT o.id, u.email AS buyer_email, st.name AS store_name,
			o.status, o.delivery_method, o.total, o.created_at
		FROM orders o
		JOIN users u ON u.id = o.buyer_id
		JOIN stores st ON st.id = o.store_id
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?`, limit, offset).Scan(&rows).Error
	return rows, total, err
}

func (r *monitorRepository) ListDeliveries(ctx context.Context, limit, offset int) ([]domain.DeliveryRow, int64, error) {
	statuses := []string{orderdomain.StatusMenungguKirim, orderdomain.StatusDikirim, orderdomain.StatusSelesai}

	var total int64
	if err := r.db.WithContext(ctx).Table("orders").Where("status IN ?", statuses).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []domain.DeliveryRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT o.id AS order_id, st.name AS store_name,
			COALESCE(du.email, '') AS driver_email,
			o.status, o.delivery_fee, o.driver_earning AS earning
		FROM orders o
		JOIN stores st ON st.id = o.store_id
		LEFT JOIN users du ON du.id = o.driver_id
		WHERE o.status IN ?
		ORDER BY o.updated_at DESC
		LIMIT ? OFFSET ?`, statuses, limit, offset).Scan(&rows).Error
	return rows, total, err
}
