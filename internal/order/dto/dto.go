package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/internal/order/usecase"
)

type PreviewRequest struct {
	DeliveryMethod string `json:"delivery_method" binding:"required,oneof=instant next_day regular"`
	DiscountCode   string `json:"discount_code" binding:"omitempty,max=50"`
}

type CheckoutRequest struct {
	AddressID      string `json:"address_id" binding:"required,uuid"`
	DeliveryMethod string `json:"delivery_method" binding:"required,oneof=instant next_day regular"`
	DiscountCode   string `json:"discount_code" binding:"omitempty,max=50"`
}

type SummaryResponse struct {
	Subtotal     int64  `json:"subtotal"`
	Discount     int64  `json:"discount"`
	DiscountCode string `json:"discount_code,omitempty"`
	DiscountKind string `json:"discount_kind,omitempty"`
	DeliveryFee  int64  `json:"delivery_fee"`
	Tax          int64  `json:"tax"`
	TaxPercent   int    `json:"tax_percent"`
	Total        int64  `json:"total"`
}

type BuyerReportResponse struct {
	TotalOrders   int            `json:"total_orders"`
	TotalSpent    int64          `json:"total_spent"`
	TotalRefunded int64          `json:"total_refunded"`
	CountByStatus map[string]int `json:"count_by_status"`
}

type SellerReportResponse struct {
	TotalOrders     int            `json:"total_orders"`
	CompletedOrders int            `json:"completed_orders"`
	TotalRevenue    int64          `json:"total_revenue"`
	TotalRefunded   int64          `json:"total_refunded"`
	CountByStatus   map[string]int `json:"count_by_status"`
}

type OrderItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int64  `json:"price"`
	Quantity    int    `json:"quantity"`
	Subtotal    int64  `json:"subtotal"`
}

type StatusHistoryResponse struct {
	Status    string `json:"status"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

type OrderResponse struct {
	ID             string                  `json:"id"`
	StoreID        string                  `json:"store_id"`
	StoreName      string                  `json:"store_name"`
	BuyerEmail     string                  `json:"buyer_email,omitempty"`
	RecipientName  string                  `json:"recipient_name"`
	Phone          string                  `json:"phone"`
	FullAddress    string                  `json:"full_address"`
	DeliveryMethod string                  `json:"delivery_method"`
	Subtotal       int64                   `json:"subtotal"`
	Discount       int64                   `json:"discount"`
	DiscountCode   string                  `json:"discount_code,omitempty"`
	DeliveryFee    int64                   `json:"delivery_fee"`
	Tax            int64                   `json:"tax"`
	TaxPercent     int                     `json:"tax_percent"`
	Total          int64                   `json:"total"`
	Status         string                  `json:"status"`
	Items          []OrderItemResponse     `json:"items"`
	StatusHistory  []StatusHistoryResponse `json:"status_history"`
	CreatedAt      string                  `json:"created_at"`
}

func ToSummaryResponse(s *usecase.CheckoutSummary) SummaryResponse {
	return SummaryResponse{
		Subtotal:     s.Subtotal,
		Discount:     s.Discount,
		DiscountCode: s.DiscountCode,
		DiscountKind: s.DiscountKind,
		DeliveryFee:  s.DeliveryFee,
		Tax:          s.Tax,
		TaxPercent:   domain.TaxRatePercent,
		Total:        s.Total,
	}
}

func ToBuyerReportResponse(r *domain.BuyerReport) BuyerReportResponse {
	return BuyerReportResponse{
		TotalOrders:   r.TotalOrders,
		TotalSpent:    r.TotalSpent,
		TotalRefunded: r.TotalRefunded,
		CountByStatus: r.CountByStatus,
	}
}

func ToSellerReportResponse(r *domain.SellerReport) SellerReportResponse {
	return SellerReportResponse{
		TotalOrders:     r.TotalOrders,
		CompletedOrders: r.CompletedOrders,
		TotalRevenue:    r.TotalRevenue,
		TotalRefunded:   r.TotalRefunded,
		CountByStatus:   r.CountByStatus,
	}
}

func ToOrderResponse(o *domain.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		})
	}
	history := make([]StatusHistoryResponse, 0, len(o.StatusHistory))
	for _, h := range o.StatusHistory {
		history = append(history, StatusHistoryResponse{
			Status:    h.Status,
			Note:      h.Note,
			CreatedAt: h.CreatedAt.Format(time.RFC3339),
		})
	}
	return OrderResponse{
		ID:             o.ID,
		StoreID:        o.StoreID,
		StoreName:      o.StoreName,
		BuyerEmail:     o.BuyerEmail,
		RecipientName:  o.RecipientName,
		Phone:          o.Phone,
		FullAddress:    o.FullAddress,
		DeliveryMethod: o.DeliveryMethod,
		Subtotal:       o.Subtotal,
		Discount:       o.Discount,
		DiscountCode:   o.DiscountCode,
		DeliveryFee:    o.DeliveryFee,
		Tax:            o.Tax,
		TaxPercent:     domain.TaxRatePercent,
		Total:          o.Total,
		Status:         o.Status,
		Items:          items,
		StatusHistory:  history,
		CreatedAt:      o.CreatedAt.Format(time.RFC3339),
	}
}

func ToOrderResponseList(list []domain.Order) []OrderResponse {
	out := make([]OrderResponse, 0, len(list))
	for i := range list {
		out = append(out, ToOrderResponse(&list[i]))
	}
	return out
}
