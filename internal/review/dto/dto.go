package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/review/domain"
)

type CreateReviewRequest struct {
	ReviewerName string `json:"reviewer_name" binding:"required,min=1,max=100"`
	Rating       int    `json:"rating" binding:"required,gte=1,lte=5"`
	Comment      string `json:"comment" binding:"required,min=1,max=1000"`
}

type ReviewResponse struct {
	ID           string `json:"id"`
	ReviewerName string `json:"reviewer_name"`
	Rating       int    `json:"rating"`
	Comment      string `json:"comment"`
	CreatedAt    string `json:"created_at"`
}

func ToReviewResponse(r *domain.AppReview) ReviewResponse {
	return ReviewResponse{
		ID:           r.ID,
		ReviewerName: r.ReviewerName,
		Rating:       r.Rating,
		Comment:      r.Comment,
		CreatedAt:    r.CreatedAt.Format(time.RFC3339),
	}
}

func ToReviewResponseList(list []domain.AppReview) []ReviewResponse {
	out := make([]ReviewResponse, 0, len(list))
	for i := range list {
		out = append(out, ToReviewResponse(&list[i]))
	}
	return out
}
