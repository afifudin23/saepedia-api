package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/discount/domain"
)

// GenerateRequest dipakai untuk generate voucher maupun promo.
// usage_limit hanya berlaku untuk voucher (diabaikan untuk promo).
type GenerateRequest struct {
	Code          string `json:"code" binding:"required,min=3,max=50"`
	DiscountType  string `json:"discount_type" binding:"required,oneof=percent fixed"`
	DiscountValue int64  `json:"discount_value" binding:"required,gt=0"`
	MaxDiscount   int64  `json:"max_discount" binding:"omitempty,gte=0"`
	MinSpend      int64  `json:"min_spend" binding:"omitempty,gte=0"`
	ExpiresAt     string `json:"expires_at" binding:"required"` // RFC3339, mis. 2026-12-31T23:59:59Z
	UsageLimit    *int   `json:"usage_limit" binding:"omitempty,gt=0"`
}

func (r GenerateRequest) ParseExpiresAt() (time.Time, error) {
	return time.Parse(time.RFC3339, r.ExpiresAt)
}

type DiscountResponse struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Kind          string `json:"kind"`
	DiscountType  string `json:"discount_type"`
	DiscountValue int64  `json:"discount_value"`
	MaxDiscount   int64  `json:"max_discount"`
	MinSpend      int64  `json:"min_spend"`
	ExpiresAt     string `json:"expires_at"`
	UsageLimit    *int   `json:"usage_limit,omitempty"`
	UsedCount     int    `json:"used_count"`
	Remaining     *int   `json:"remaining_usage,omitempty"`
	CreatedAt     string `json:"created_at"`
}

func ToDiscountResponse(d *domain.Discount) DiscountResponse {
	res := DiscountResponse{
		ID:            d.ID,
		Code:          d.Code,
		Kind:          d.Kind,
		DiscountType:  d.DiscountType,
		DiscountValue: d.DiscountValue,
		MaxDiscount:   d.MaxDiscount,
		MinSpend:      d.MinSpend,
		ExpiresAt:     d.ExpiresAt.Format(time.RFC3339),
		UsageLimit:    d.UsageLimit,
		UsedCount:     d.UsedCount,
		CreatedAt:     d.CreatedAt.Format(time.RFC3339),
	}
	if d.Kind == domain.KindVoucher && d.UsageLimit != nil {
		rem := *d.UsageLimit - d.UsedCount
		if rem < 0 {
			rem = 0
		}
		res.Remaining = &rem
	}
	return res
}

func ToDiscountResponseList(list []domain.Discount) []DiscountResponse {
	out := make([]DiscountResponse, 0, len(list))
	for i := range list {
		out = append(out, ToDiscountResponse(&list[i]))
	}
	return out
}
