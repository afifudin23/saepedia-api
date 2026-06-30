package domain

import (
	"context"
	"time"
)

// AppReview adalah review tentang aplikasi/website SEAPEDIA (bukan review produk).
// Boleh diisi guest tanpa checkout/transaksi.
type AppReview struct {
	ID           string
	ReviewerName string
	Rating       int
	Comment      string
	CreatedAt    time.Time
}

type ReviewRepository interface {
	Create(ctx context.Context, r *AppReview) error
	List(ctx context.Context, limit, offset int) ([]AppReview, int64, error)
}
