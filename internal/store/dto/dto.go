package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/store/domain"
)

type UpsertStoreRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=150"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

type StoreResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// PublicStoreSummary dipakai untuk endpoint publik (buyer/guest).
type PublicStoreSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ToStoreResponse(s *domain.Store) StoreResponse {
	return StoreResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPublicStoreSummary(s *domain.Store) PublicStoreSummary {
	return PublicStoreSummary{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
	}
}

func ToPublicStoreSummaryList(list []domain.Store) []PublicStoreSummary {
	out := make([]PublicStoreSummary, 0, len(list))
	for i := range list {
		out = append(out, ToPublicStoreSummary(&list[i]))
	}
	return out
}
