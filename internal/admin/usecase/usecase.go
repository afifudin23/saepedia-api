package usecase

import (
	"context"
	"time"

	admindomain "github.com/afifudin23/saepedia-api/internal/admin/domain"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	orderusecase "github.com/afifudin23/saepedia-api/internal/order/usecase"
	"github.com/afifudin23/saepedia-api/pkg/clock"
)

// TimeSettingStore mempersist offset simulasi waktu.
type TimeSettingStore interface {
	GetTimeOffset(ctx context.Context) (time.Duration, error)
	SetTimeOffset(ctx context.Context, d time.Duration) error
}

// SimulateResult adalah hasil memajukan waktu + overdue yang ikut diproses.
type SimulateResult struct {
	Now            time.Time
	OffsetDays     float64
	ProcessedCount int
	Processed      []orderdomain.Order
}

type AdminUsecase interface {
	Summary(ctx context.Context) (*admindomain.Summary, error)
	ListUsers(ctx context.Context, limit, offset int) ([]admindomain.UserRow, int64, error)
	ListStores(ctx context.Context, limit, offset int) ([]admindomain.StoreRow, int64, error)
	ListProducts(ctx context.Context, limit, offset int) ([]admindomain.ProductRow, int64, error)
	ListOrders(ctx context.Context, limit, offset int) ([]admindomain.OrderRow, int64, error)
	ListDeliveries(ctx context.Context, limit, offset int) ([]admindomain.DeliveryRow, int64, error)
	ListOverdue(ctx context.Context) ([]orderdomain.Order, error)

	Now() time.Time
	AdvanceDays(ctx context.Context, days int) (*SimulateResult, error)
	RunOverdue(ctx context.Context) ([]orderdomain.Order, error)
}

type adminUsecase struct {
	repo     admindomain.MonitorRepository
	orderUC  orderusecase.OrderUsecase
	settings TimeSettingStore
}

func New(repo admindomain.MonitorRepository, orderUC orderusecase.OrderUsecase, settings TimeSettingStore) AdminUsecase {
	return &adminUsecase{repo: repo, orderUC: orderUC, settings: settings}
}

func (uc *adminUsecase) Summary(ctx context.Context) (*admindomain.Summary, error) {
	s, err := uc.repo.Summary(ctx)
	if err != nil {
		return nil, err
	}
	overdue, err := uc.orderUC.ListOverdue(ctx)
	if err != nil {
		return nil, err
	}
	s.OverdueOrders = int64(len(overdue))
	return s, nil
}

func (uc *adminUsecase) ListUsers(ctx context.Context, limit, offset int) ([]admindomain.UserRow, int64, error) {
	return uc.repo.ListUsers(ctx, limit, offset)
}

func (uc *adminUsecase) ListStores(ctx context.Context, limit, offset int) ([]admindomain.StoreRow, int64, error) {
	return uc.repo.ListStores(ctx, limit, offset)
}

func (uc *adminUsecase) ListProducts(ctx context.Context, limit, offset int) ([]admindomain.ProductRow, int64, error) {
	return uc.repo.ListProducts(ctx, limit, offset)
}

func (uc *adminUsecase) ListOrders(ctx context.Context, limit, offset int) ([]admindomain.OrderRow, int64, error) {
	return uc.repo.ListOrders(ctx, limit, offset)
}

func (uc *adminUsecase) ListDeliveries(ctx context.Context, limit, offset int) ([]admindomain.DeliveryRow, int64, error) {
	return uc.repo.ListDeliveries(ctx, limit, offset)
}

func (uc *adminUsecase) ListOverdue(ctx context.Context) ([]orderdomain.Order, error) {
	return uc.orderUC.ListOverdue(ctx)
}

func (uc *adminUsecase) Now() time.Time { return clock.Now() }

// AdvanceDays memajukan waktu virtual N hari, mempersist offset, lalu langsung
// memproses order yang menjadi overdue akibat lompatan waktu tersebut.
func (uc *adminUsecase) AdvanceDays(ctx context.Context, days int) (*SimulateResult, error) {
	clock.Advance(time.Duration(days) * 24 * time.Hour)
	if err := uc.settings.SetTimeOffset(ctx, clock.Offset()); err != nil {
		return nil, err
	}

	processed, err := uc.orderUC.RunOverdue(ctx)
	if err != nil {
		return nil, err
	}
	return &SimulateResult{
		Now:            clock.Now(),
		OffsetDays:     clock.Offset().Hours() / 24,
		ProcessedCount: len(processed),
		Processed:      processed,
	}, nil
}

func (uc *adminUsecase) RunOverdue(ctx context.Context) ([]orderdomain.Order, error) {
	return uc.orderUC.RunOverdue(ctx)
}
