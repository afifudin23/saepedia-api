package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/delivery/domain"
	"github.com/afifudin23/saepedia-api/internal/delivery/usecase"
)

type JobResponse struct {
	OrderID        string `json:"order_id"`
	StoreName      string `json:"store_name"`
	RecipientName  string `json:"recipient_name"`
	FullAddress    string `json:"full_address"`
	DeliveryMethod string `json:"delivery_method"`
	DeliveryFee    int64  `json:"delivery_fee"`
	Earning        int64  `json:"earning"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}

type DashboardResponse struct {
	ActiveJobs     []JobResponse `json:"active_jobs"`
	History        []JobResponse `json:"history"`
	TotalEarnings  int64         `json:"total_earnings"`
	CompletedCount int           `json:"completed_count"`
}

func ToJobResponse(j *domain.Job) JobResponse {
	return JobResponse{
		OrderID:        j.OrderID,
		StoreName:      j.StoreName,
		RecipientName:  j.RecipientName,
		FullAddress:    j.FullAddress,
		DeliveryMethod: j.DeliveryMethod,
		DeliveryFee:    j.DeliveryFee,
		Earning:        j.Earning,
		Status:         j.Status,
		CreatedAt:      j.CreatedAt.Format(time.RFC3339),
	}
}

func ToJobResponseList(list []domain.Job) []JobResponse {
	out := make([]JobResponse, 0, len(list))
	for i := range list {
		out = append(out, ToJobResponse(&list[i]))
	}
	return out
}

func ToDashboardResponse(d *usecase.Dashboard) DashboardResponse {
	return DashboardResponse{
		ActiveJobs:     ToJobResponseList(d.ActiveJobs),
		History:        ToJobResponseList(d.History),
		TotalEarnings:  d.TotalEarnings,
		CompletedCount: d.CompletedCount,
	}
}
