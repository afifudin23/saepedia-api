package repository

import (
	"context"
	"time"

	"github.com/afifudin23/saepedia-api/internal/delivery/domain"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
)

// jobRow menampung hasil join orders + stores.
type jobRow struct {
	OrderID        string
	StoreName      string
	RecipientName  string
	FullAddress    string
	DeliveryMethod string
	DeliveryFee    int64
	DriverEarning  int64
	Status         string
	CreatedAt      time.Time
}

func (r jobRow) toDomain() domain.Job {
	earning := orderdomain.CalcDriverEarning(r.DeliveryFee)
	if r.Status == orderdomain.StatusSelesai {
		earning = r.DriverEarning // pakai nilai aktual yang tersimpan saat selesai
	}
	return domain.Job{
		OrderID:        r.OrderID,
		StoreName:      r.StoreName,
		RecipientName:  r.RecipientName,
		FullAddress:    r.FullAddress,
		DeliveryMethod: r.DeliveryMethod,
		DeliveryFee:    r.DeliveryFee,
		Earning:        earning,
		Status:         r.Status,
		CreatedAt:      r.CreatedAt,
	}
}

type deliveryRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.DeliveryRepository {
	return &deliveryRepository{db: db}
}

const jobSelect = `orders.id AS order_id, stores.name AS store_name, orders.recipient_name,
	orders.full_address, orders.delivery_method, orders.delivery_fee, orders.driver_earning,
	orders.status, orders.created_at`

func (r *deliveryRepository) base(ctx context.Context) *gorm.DB {
	return tx.DB(ctx, r.db).
		Table("orders").
		Joins("JOIN stores ON stores.id = orders.store_id")
}

func (r *deliveryRepository) ListAvailable(ctx context.Context, limit, offset int) ([]domain.Job, int64, error) {
	cond := "orders.status = ? AND orders.driver_id IS NULL"

	var total int64
	if err := r.base(ctx).Where(cond, orderdomain.StatusMenungguKirim).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []jobRow
	err := r.base(ctx).
		Select(jobSelect).
		Where(cond, orderdomain.StatusMenungguKirim).
		Order("orders.created_at ASC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return toJobs(rows), total, nil
}

func (r *deliveryRepository) GetAvailable(ctx context.Context, orderID string) (*domain.Job, error) {
	var row jobRow
	err := r.base(ctx).
		Select(jobSelect).
		Where("orders.id = ? AND orders.status = ? AND orders.driver_id IS NULL", orderID, orderdomain.StatusMenungguKirim).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.OrderID == "" {
		return nil, domain.ErrJobNotFound
	}
	job := row.toDomain()
	return &job, nil
}

// OrderState membaca status order untuk dicek usecase sebelum transisi.
func (r *deliveryRepository) OrderState(ctx context.Context, orderID string) (domain.OrderState, error) {
	var row struct {
		Status      string
		DriverID    *string
		DeliveryFee int64
	}
	err := tx.DB(ctx, r.db).
		Table("orders").
		Select("status, driver_id, delivery_fee").
		Where("id = ?", orderID).
		Scan(&row).Error
	if err != nil {
		return domain.OrderState{}, err
	}
	if row.Status == "" {
		return domain.OrderState{Found: false}, nil
	}
	state := domain.OrderState{Found: true, Status: row.Status, DeliveryFee: row.DeliveryFee}
	if row.DriverID != nil {
		state.DriverID = *row.DriverID
	}
	return state, nil
}

// AssignDriver: set driver + ubah status secara atomik (anti dua driver).
func (r *deliveryRepository) AssignDriver(ctx context.Context, driverID, orderID, from, to string) (bool, error) {
	res := tx.DB(ctx, r.db).Exec(
		"UPDATE orders SET driver_id = ?, status = ?, taken_at = NOW(), updated_at = NOW() WHERE id = ? AND status = ? AND driver_id IS NULL",
		driverID, to, orderID, from,
	)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// CompleteJob: ubah status + simpan earning (dihitung usecase) secara atomik.
func (r *deliveryRepository) CompleteJob(ctx context.Context, driverID, orderID, from, to string, earning int64) (bool, error) {
	res := tx.DB(ctx, r.db).Exec(
		"UPDATE orders SET status = ?, driver_earning = ?, completed_at = NOW(), updated_at = NOW() WHERE id = ? AND driver_id = ? AND status = ?",
		to, earning, orderID, driverID, from,
	)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *deliveryRepository) AddHistory(ctx context.Context, orderID, status, note string) error {
	return tx.DB(ctx, r.db).Exec(
		"INSERT INTO order_status_histories (order_id, status, note) VALUES (?, ?, ?)",
		orderID, status, note,
	).Error
}

func (r *deliveryRepository) ListByDriver(ctx context.Context, driverID string, statuses []string, limit, offset int) ([]domain.Job, int64, error) {
	cond := "orders.driver_id = ? AND orders.status IN ?"

	var total int64
	if err := r.base(ctx).Where(cond, driverID, statuses).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []jobRow
	err := r.base(ctx).
		Select(jobSelect).
		Where(cond, driverID, statuses).
		Order("orders.updated_at DESC").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return toJobs(rows), total, nil
}

func (r *deliveryRepository) Earnings(ctx context.Context, driverID string) (int64, int, error) {
	var agg struct {
		Total     int64
		Completed int
	}
	err := tx.DB(ctx, r.db).
		Table("orders").
		Select("COALESCE(SUM(driver_earning), 0) AS total, COUNT(*) AS completed").
		Where("driver_id = ? AND status = ?", driverID, orderdomain.StatusSelesai).
		Scan(&agg).Error
	if err != nil {
		return 0, 0, err
	}
	return agg.Total, agg.Completed, nil
}

func (r *deliveryRepository) GetByDriver(ctx context.Context, driverID, orderID string) (*domain.Job, error) {
	var row jobRow
	err := r.base(ctx).
		Select(jobSelect).
		Where("orders.id = ? AND orders.driver_id = ?", orderID, driverID).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.OrderID == "" {
		return nil, domain.ErrJobNotFound
	}
	job := row.toDomain()
	return &job, nil
}

func toJobs(rows []jobRow) []domain.Job {
	out := make([]domain.Job, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomain())
	}
	return out
}
