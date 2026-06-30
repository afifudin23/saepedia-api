package usecase

import (
	"context"

	"github.com/afifudin23/saepedia-api/internal/delivery/domain"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
)

// Dashboard merangkum job aktif, riwayat, dan pendapatan driver.
type Dashboard struct {
	ActiveJobs     []domain.Job
	History        []domain.Job
	TotalEarnings  int64
	CompletedCount int
}

type DeliveryUsecase interface {
	AvailableJobs(ctx context.Context, limit, offset int) ([]domain.Job, int64, error)
	JobDetail(ctx context.Context, orderID string) (*domain.Job, error)
	Take(ctx context.Context, driverID, orderID string) (*domain.Job, error)
	Complete(ctx context.Context, driverID, orderID string) (*domain.Job, error)
	Dashboard(ctx context.Context, driverID string) (*Dashboard, error)
}

type deliveryUsecase struct {
	repo  domain.DeliveryRepository
	txMgr *tx.Manager
}

func New(repo domain.DeliveryRepository, txMgr *tx.Manager) DeliveryUsecase {
	return &deliveryUsecase{repo: repo, txMgr: txMgr}
}

func (uc *deliveryUsecase) AvailableJobs(ctx context.Context, limit, offset int) ([]domain.Job, int64, error) {
	return uc.repo.ListAvailable(ctx, limit, offset)
}

func (uc *deliveryUsecase) JobDetail(ctx context.Context, orderID string) (*domain.Job, error) {
	return uc.repo.GetAvailable(ctx, orderID)
}

// Take: aturan "hanya job Menunggu Pengirim & belum diambil" diputuskan di usecase;
// repo hanya melakukan update atomik (guard) untuk cegah balapan dua driver.
func (uc *deliveryUsecase) Take(ctx context.Context, driverID, orderID string) (*domain.Job, error) {
	err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
		state, err := uc.repo.OrderState(ctx, orderID)
		if err != nil {
			return err
		}
		if !state.Found {
			return domain.ErrJobNotFound
		}
		if state.DriverID != "" {
			return domain.ErrJobTaken
		}
		if state.Status != orderdomain.StatusMenungguKirim {
			return domain.ErrJobInvalidState
		}

		ok, err := uc.repo.AssignDriver(ctx, driverID, orderID, orderdomain.StatusMenungguKirim, orderdomain.StatusDikirim)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrJobTaken // keburu diambil driver lain
		}
		return uc.repo.AddHistory(ctx, orderID, orderdomain.StatusDikirim, "driver took the job, package on the way")
	})
	if err != nil {
		return nil, err
	}
	return uc.repo.GetByDriver(ctx, driverID, orderID)
}

// Complete: transisi & perhitungan earning (80% ongkir) diputuskan di usecase.
func (uc *deliveryUsecase) Complete(ctx context.Context, driverID, orderID string) (*domain.Job, error) {
	err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
		state, err := uc.repo.OrderState(ctx, orderID)
		if err != nil {
			return err
		}
		if !state.Found {
			return domain.ErrJobNotFound
		}
		if state.DriverID != driverID {
			return domain.ErrJobNotYours
		}
		if state.Status != orderdomain.StatusDikirim {
			return domain.ErrJobInvalidState
		}

		earning := orderdomain.CalcDriverEarning(state.DeliveryFee)
		ok, err := uc.repo.CompleteJob(ctx, driverID, orderID, orderdomain.StatusDikirim, orderdomain.StatusSelesai, earning)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrJobInvalidState
		}
		return uc.repo.AddHistory(ctx, orderID, orderdomain.StatusSelesai, "driver confirmed delivery completed")
	})
	if err != nil {
		return nil, err
	}
	return uc.repo.GetByDriver(ctx, driverID, orderID)
}

func (uc *deliveryUsecase) Dashboard(ctx context.Context, driverID string) (*Dashboard, error) {
	active, _, err := uc.repo.ListByDriver(ctx, driverID, []string{orderdomain.StatusDikirim}, 50, 0)
	if err != nil {
		return nil, err
	}
	history, _, err := uc.repo.ListByDriver(ctx, driverID, []string{orderdomain.StatusSelesai}, 50, 0)
	if err != nil {
		return nil, err
	}
	total, completed, err := uc.repo.Earnings(ctx, driverID)
	if err != nil {
		return nil, err
	}
	return &Dashboard{
		ActiveJobs:     active,
		History:        history,
		TotalEarnings:  total,
		CompletedCount: completed,
	}, nil
}
