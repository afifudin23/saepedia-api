package repository

import (
	"context"
	"errors"

	"github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type orderRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

// Create menyimpan order + item-nya. Status & nominal sudah dihitung usecase.
func (r *orderRepository) Create(ctx context.Context, o *domain.Order) error {
	db := tx.DB(ctx, r.db)

	model := OrderModel{
		BuyerID:        o.BuyerID,
		StoreID:        o.StoreID,
		RecipientName:  o.RecipientName,
		Phone:          o.Phone,
		FullAddress:    o.FullAddress,
		DeliveryMethod: o.DeliveryMethod,
		Subtotal:       o.Subtotal,
		Discount:       o.Discount,
		DiscountCode:   o.DiscountCode,
		DeliveryFee:    o.DeliveryFee,
		Tax:            o.Tax,
		Total:          o.Total,
		Status:         o.Status,
	}
	if err := db.Create(&model).Error; err != nil {
		return err
	}

	for _, it := range o.Items {
		if err := db.Create(&OrderItemModel{
			OrderID:     model.ID,
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		}).Error; err != nil {
			return err
		}
	}

	o.ID = model.ID
	o.CreatedAt = model.CreatedAt
	o.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *orderRepository) AddStatusHistory(ctx context.Context, orderID, status, note string) error {
	return tx.DB(ctx, r.db).Create(&OrderStatusHistoryModel{
		OrderID: orderID, Status: status, Note: note,
	}).Error
}

func (r *orderRepository) UpdateStatusGuarded(ctx context.Context, orderID, from, to string) (bool, error) {
	res := tx.DB(ctx, r.db).Exec(
		"UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ? AND status = ?",
		to, orderID, from,
	)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *orderRepository) MarkRefunded(ctx context.Context, orderID string) error {
	return tx.DB(ctx, r.db).Exec(
		"UPDATE orders SET status = ?, refunded_at = NOW(), updated_at = NOW() WHERE id = ?",
		domain.StatusDikembalikan, orderID,
	).Error
}

func (r *orderRepository) GetForUpdate(ctx context.Context, orderID string) (*domain.Order, error) {
	var model OrderModel
	err := tx.DB(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", orderID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *orderRepository) Items(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	var models []OrderItemModel
	if err := tx.DB(ctx, r.db).Where("order_id = ?", orderID).Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.OrderItem, 0, len(models))
	for _, it := range models {
		out = append(out, domain.OrderItem{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		})
	}
	return out, nil
}

func (r *orderRepository) FindByIDForBuyer(ctx context.Context, buyerID, orderID string) (*domain.Order, error) {
	return r.findOneWhere(ctx, "id = ? AND buyer_id = ?", orderID, buyerID)
}

func (r *orderRepository) FindByIDForStore(ctx context.Context, storeID, orderID string) (*domain.Order, error) {
	return r.findOneWhere(ctx, "id = ? AND store_id = ?", orderID, storeID)
}

func (r *orderRepository) findOneWhere(ctx context.Context, query string, args ...any) (*domain.Order, error) {
	var model OrderModel
	err := tx.DB(ctx, r.db).
		Preload("Items").
		Preload("History", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Where(query, args...).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	order := model.toDomain()
	r.enrich(ctx, []*domain.Order{order})
	return order, nil
}

func (r *orderRepository) ListByBuyer(ctx context.Context, buyerID string, limit, offset int) ([]domain.Order, int64, error) {
	return r.list(ctx, "buyer_id = ?", buyerID, limit, offset)
}

func (r *orderRepository) ListByStore(ctx context.Context, storeID string, limit, offset int) ([]domain.Order, int64, error) {
	return r.list(ctx, "store_id = ?", storeID, limit, offset)
}

func (r *orderRepository) list(ctx context.Context, where string, arg any, limit, offset int) ([]domain.Order, int64, error) {
	var models []OrderModel
	var total int64

	if err := tx.DB(ctx, r.db).Model(&OrderModel{}).Where(where, arg).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.DB(ctx, r.db).
		Preload("Items").
		Preload("History", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Where(where, arg).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	orders := make([]domain.Order, 0, len(models))
	for _, m := range models {
		orders = append(orders, *m.toDomain())
	}
	ptrs := make([]*domain.Order, 0, len(orders))
	for i := range orders {
		ptrs = append(ptrs, &orders[i])
	}
	r.enrich(ctx, ptrs)
	return orders, total, nil
}

func (r *orderRepository) ListRefundCandidates(ctx context.Context) ([]domain.Order, error) {
	var models []OrderModel
	err := tx.DB(ctx, r.db).
		Where("status IN ? AND refunded_at IS NULL",
			[]string{domain.StatusDikemas, domain.StatusMenungguKirim, domain.StatusDikirim}).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Order, 0, len(models))
	for _, m := range models {
		out = append(out, *m.toDomain())
	}
	return out, nil
}

func (r *orderRepository) BuyerReport(ctx context.Context, buyerID string) (*domain.BuyerReport, error) {
	report := &domain.BuyerReport{CountByStatus: map[string]int{}}

	var agg struct {
		TotalOrders   int
		TotalSpent    int64
		TotalRefunded int64
	}
	err := tx.DB(ctx, r.db).Raw(`
		SELECT
			COUNT(*) AS total_orders,
			COALESCE(SUM(CASE WHEN status <> ? THEN total ELSE 0 END), 0) AS total_spent,
			COALESCE(SUM(CASE WHEN status = ? THEN total ELSE 0 END), 0) AS total_refunded
		FROM orders WHERE buyer_id = ?`,
		domain.StatusDikembalikan, domain.StatusDikembalikan, buyerID,
	).Scan(&agg).Error
	if err != nil {
		return nil, err
	}
	report.TotalOrders = agg.TotalOrders
	report.TotalSpent = agg.TotalSpent
	report.TotalRefunded = agg.TotalRefunded

	if err := r.countByStatus(ctx, "buyer_id = ?", buyerID, report.CountByStatus); err != nil {
		return nil, err
	}
	return report, nil
}

func (r *orderRepository) SellerReport(ctx context.Context, storeID string) (*domain.SellerReport, error) {
	report := &domain.SellerReport{CountByStatus: map[string]int{}}

	var agg struct {
		TotalOrders     int
		CompletedOrders int
		TotalRevenue    int64
		TotalRefunded   int64
	}
	err := tx.DB(ctx, r.db).Raw(`
		SELECT
			COUNT(*) AS total_orders,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS completed_orders,
			COALESCE(SUM(CASE WHEN status = ? THEN (subtotal - discount) ELSE 0 END), 0) AS total_revenue,
			COALESCE(SUM(CASE WHEN status = ? THEN total ELSE 0 END), 0) AS total_refunded
		FROM orders WHERE store_id = ?`,
		domain.StatusSelesai, domain.StatusSelesai, domain.StatusDikembalikan, storeID,
	).Scan(&agg).Error
	if err != nil {
		return nil, err
	}
	report.TotalOrders = agg.TotalOrders
	report.CompletedOrders = agg.CompletedOrders
	report.TotalRevenue = agg.TotalRevenue
	report.TotalRefunded = agg.TotalRefunded

	if err := r.countByStatus(ctx, "store_id = ?", storeID, report.CountByStatus); err != nil {
		return nil, err
	}
	return report, nil
}

func (r *orderRepository) countByStatus(ctx context.Context, where string, arg any, out map[string]int) error {
	var rows []struct {
		Status string
		Count  int
	}
	err := tx.DB(ctx, r.db).
		Table("orders").
		Select("status, COUNT(*) AS count").
		Where(where, arg).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return nil
}

// enrich mengisi StoreName & BuyerEmail lewat query batch (hindari N+1).
func (r *orderRepository) enrich(ctx context.Context, orders []*domain.Order) {
	if len(orders) == 0 {
		return
	}

	storeIDs := make([]string, 0, len(orders))
	buyerIDs := make([]string, 0, len(orders))
	for _, o := range orders {
		storeIDs = append(storeIDs, o.StoreID)
		buyerIDs = append(buyerIDs, o.BuyerID)
	}

	storeNames := r.namesFor(ctx, "stores", "name", storeIDs)
	buyerEmails := r.namesFor(ctx, "users", "email", buyerIDs)

	for _, o := range orders {
		o.StoreName = storeNames[o.StoreID]
		o.BuyerEmail = buyerEmails[o.BuyerID]
	}
}

type idName struct {
	ID   string
	Name string
}

func (r *orderRepository) namesFor(ctx context.Context, table, col string, ids []string) map[string]string {
	out := make(map[string]string)
	if len(ids) == 0 {
		return out
	}
	var rows []idName
	tx.DB(ctx, r.db).
		Table(table).
		Select("id, "+col+" AS name").
		Where("id IN ?", ids).
		Scan(&rows)
	for _, row := range rows {
		out[row.ID] = row.Name
	}
	return out
}
