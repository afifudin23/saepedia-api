package dto

import (
	"strings"
	"time"

	"github.com/afifudin23/saepedia-api/internal/admin/domain"
)

type SummaryResponse struct {
	Users            int64            `json:"users"`
	Stores           int64            `json:"stores"`
	Products         int64            `json:"products"`
	Orders           int64            `json:"orders"`
	OrdersByStatus   map[string]int64 `json:"orders_by_status"`
	Vouchers         int64            `json:"vouchers"`
	Promos           int64            `json:"promos"`
	AvailableJobs    int64            `json:"available_delivery_jobs"`
	ActiveDeliveries int64            `json:"active_deliveries"`
	OverdueOrders    int64            `json:"overdue_orders"`
}

func ToSummaryResponse(s *domain.Summary) SummaryResponse {
	return SummaryResponse{
		Users:            s.Users,
		Stores:           s.Stores,
		Products:         s.Products,
		Orders:           s.Orders,
		OrdersByStatus:   s.OrdersByStatus,
		Vouchers:         s.Vouchers,
		Promos:           s.Promos,
		AvailableJobs:    s.AvailableJobs,
		ActiveDeliveries: s.ActiveDeliveries,
		OverdueOrders:    s.OverdueOrders,
	}
}

type UserResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	IsAdmin   bool     `json:"is_admin"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

func ToUserResponseList(rows []domain.UserRow) []UserResponse {
	out := make([]UserResponse, 0, len(rows))
	for _, r := range rows {
		roles := []string{}
		if r.Roles != "" {
			roles = strings.Split(r.Roles, ",")
		}
		out = append(out, UserResponse{
			ID: r.ID, Email: r.Email, IsAdmin: r.IsAdmin,
			Roles: roles, CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

type StoreResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	ProductCount int64  `json:"product_count"`
	CreatedAt    string `json:"created_at"`
}

func ToStoreResponseList(rows []domain.StoreRow) []StoreResponse {
	out := make([]StoreResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, StoreResponse{
			ID: r.ID, Name: r.Name, Owner: r.Owner,
			ProductCount: r.ProductCount, CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

type ProductResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StoreName string `json:"store_name"`
	Price     int64  `json:"price"`
	Stock     int    `json:"stock"`
	CreatedAt string `json:"created_at"`
}

func ToProductResponseList(rows []domain.ProductRow) []ProductResponse {
	out := make([]ProductResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, ProductResponse{
			ID: r.ID, Name: r.Name, StoreName: r.StoreName,
			Price: r.Price, Stock: r.Stock, CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

type OrderRowResponse struct {
	ID             string `json:"id"`
	BuyerEmail     string `json:"buyer_email"`
	StoreName      string `json:"store_name"`
	Status         string `json:"status"`
	DeliveryMethod string `json:"delivery_method"`
	Total          int64  `json:"total"`
	CreatedAt      string `json:"created_at"`
}

func ToOrderRowResponseList(rows []domain.OrderRow) []OrderRowResponse {
	out := make([]OrderRowResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, OrderRowResponse{
			ID: r.ID, BuyerEmail: r.BuyerEmail, StoreName: r.StoreName,
			Status: r.Status, DeliveryMethod: r.DeliveryMethod, Total: r.Total,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

type DeliveryRowResponse struct {
	OrderID     string `json:"order_id"`
	StoreName   string `json:"store_name"`
	DriverEmail string `json:"driver_email"`
	Status      string `json:"status"`
	DeliveryFee int64  `json:"delivery_fee"`
	Earning     int64  `json:"earning"`
}

func ToDeliveryRowResponseList(rows []domain.DeliveryRow) []DeliveryRowResponse {
	out := make([]DeliveryRowResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, DeliveryRowResponse{
			OrderID: r.OrderID, StoreName: r.StoreName, DriverEmail: r.DriverEmail,
			Status: r.Status, DeliveryFee: r.DeliveryFee, Earning: r.Earning,
		})
	}
	return out
}
