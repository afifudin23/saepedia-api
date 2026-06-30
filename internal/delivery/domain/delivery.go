package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrJobNotFound     = errors.New("delivery job not found or not available")
	ErrJobTaken        = errors.New("delivery job has already been taken by another driver")
	ErrJobNotYours     = errors.New("this delivery job is not assigned to you")
	ErrJobInvalidState = errors.New("delivery job is not in a valid state for this action")
)

// Job adalah representasi pekerjaan antar dari sudut pandang driver (atas data order).
type Job struct {
	OrderID        string
	StoreName      string
	RecipientName  string
	FullAddress    string
	DeliveryMethod string
	DeliveryFee    int64
	Earning        int64 // potensi (job tersedia/aktif) atau aktual (selesai)
	Status         string
	CreatedAt      time.Time
}

// OrderState adalah snapshot status order untuk pengambilan keputusan di usecase.
type OrderState struct {
	Found       bool
	Status      string
	DriverID    string // "" bila belum ada driver
	DeliveryFee int64
}

// DeliveryRepository hanya operasi data atas order (lensa driver). Keputusan
// transisi & perhitungan earning ada di usecase.
type DeliveryRepository interface {
	ListAvailable(ctx context.Context, limit, offset int) ([]Job, int64, error)
	GetAvailable(ctx context.Context, orderID string) (*Job, error)
	ListByDriver(ctx context.Context, driverID string, statuses []string, limit, offset int) ([]Job, int64, error)
	GetByDriver(ctx context.Context, driverID, orderID string) (*Job, error)
	// Earnings mengembalikan total pendapatan + jumlah job selesai milik driver.
	Earnings(ctx context.Context, driverID string) (total int64, completed int, err error)

	// OrderState membaca status order (untuk dicek usecase sebelum transisi).
	OrderState(ctx context.Context, orderID string) (OrderState, error)
	// AssignDriver: set driver + ubah status secara atomik (false bila syarat tak terpenuhi).
	AssignDriver(ctx context.Context, driverID, orderID, from, to string) (bool, error)
	// CompleteJob: ubah status + simpan earning secara atomik (earning dihitung usecase).
	CompleteJob(ctx context.Context, driverID, orderID, from, to string, earning int64) (bool, error)
	AddHistory(ctx context.Context, orderID, status, note string) error
}
