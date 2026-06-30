package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/address/domain"
)

type UpsertAddressRequest struct {
	RecipientName string `json:"recipient_name" binding:"required,min=1,max=100"`
	Phone         string `json:"phone" binding:"required,min=5,max=20"`
	FullAddress   string `json:"full_address" binding:"required,min=5,max=500"`
	IsPrimary     bool   `json:"is_primary"`
}

type AddressResponse struct {
	ID            string `json:"id"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	FullAddress   string `json:"full_address"`
	IsPrimary     bool   `json:"is_primary"`
	CreatedAt     string `json:"created_at"`
}

func ToAddressResponse(a *domain.Address) AddressResponse {
	return AddressResponse{
		ID:            a.ID,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		FullAddress:   a.FullAddress,
		IsPrimary:     a.IsPrimary,
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
	}
}

func ToAddressResponseList(list []domain.Address) []AddressResponse {
	out := make([]AddressResponse, 0, len(list))
	for i := range list {
		out = append(out, ToAddressResponse(&list[i]))
	}
	return out
}
