package dto

import "github.com/afifudin23/saepedia-api/internal/cart/domain"

type AddItemRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type CartItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Price       int64  `json:"price"`
	Quantity    int    `json:"quantity"`
	Stock       int    `json:"stock"`
	Subtotal    int64  `json:"subtotal"`
}

type CartResponse struct {
	ID        string             `json:"id"`
	StoreID   string             `json:"store_id"`
	StoreName string             `json:"store_name"`
	Items     []CartItemResponse `json:"items"`
	Subtotal  int64              `json:"subtotal"`
	ItemCount int                `json:"item_count"`
}

func ToCartResponse(c *domain.Cart) CartResponse {
	items := make([]CartItemResponse, 0, len(c.Items))
	count := 0
	for _, it := range c.Items {
		items = append(items, CartItemResponse{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Stock:       it.Stock,
			Subtotal:    it.Subtotal,
		})
		count += it.Quantity
	}
	return CartResponse{
		ID:        c.ID,
		StoreID:   c.StoreID,
		StoreName: c.StoreName,
		Items:     items,
		Subtotal:  c.Subtotal,
		ItemCount: count,
	}
}
