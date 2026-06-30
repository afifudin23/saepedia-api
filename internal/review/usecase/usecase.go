package usecase

import (
	"context"
	"html"
	"strings"

	"github.com/afifudin23/saepedia-api/internal/review/domain"
)

type ReviewUsecase interface {
	Create(ctx context.Context, name string, rating int, comment string) (*domain.AppReview, error)
	List(ctx context.Context, limit, offset int) ([]domain.AppReview, int64, error)
}

type reviewUsecase struct {
	repo domain.ReviewRepository
}

func New(repo domain.ReviewRepository) ReviewUsecase {
	return &reviewUsecase{repo: repo}
}

func (uc *reviewUsecase) Create(ctx context.Context, name string, rating int, comment string) (*domain.AppReview, error) {
	// Sanitasi dasar: konten user disimpan ter-escape agar saat ditampilkan
	// di frontend muncul sebagai teks biasa, tidak mengeksekusi script.
	// Pencegahan XSS yang lebih formal dikerjakan di Level 7.
	review := &domain.AppReview{
		ReviewerName: html.EscapeString(strings.TrimSpace(name)),
		Rating:       rating,
		Comment:      html.EscapeString(strings.TrimSpace(comment)),
	}
	if err := uc.repo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

func (uc *reviewUsecase) List(ctx context.Context, limit, offset int) ([]domain.AppReview, int64, error) {
	return uc.repo.List(ctx, limit, offset)
}
