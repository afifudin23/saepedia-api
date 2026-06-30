package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/product/domain"
)

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"omitempty,max=2000"`
	Price       int64  `json:"price" binding:"required,gte=0"`
	Stock       int    `json:"stock" binding:"gte=0"`
	ImageURL    string `json:"image_url" binding:"omitempty,url,max=500"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"omitempty,max=2000"`
	Price       int64  `json:"price" binding:"required,gte=0"`
	Stock       int    `json:"stock" binding:"gte=0"`
	ImageURL    string `json:"image_url" binding:"omitempty,url,max=500"`
}

// ProductResponse — view seller (lengkap).
type ProductResponse struct {
	ID          string `json:"id"`
	StoreID     string `json:"store_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	ImageURL    string `json:"image_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// PublicProductResponse — view publik (dengan info toko).
type PublicProductResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	ImageURL    string `json:"image_url"`
	Store       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"store"`
	CreatedAt string `json:"created_at"`
}

func ToProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		StoreID:     p.StoreID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func ToProductResponseList(list []domain.Product) []ProductResponse {
	out := make([]ProductResponse, 0, len(list))
	for i := range list {
		out = append(out, ToProductResponse(&list[i]))
	}
	return out
}

func ToPublicProductResponse(p *domain.Product) PublicProductResponse {
	r := PublicProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
	r.Store.ID = p.StoreID
	r.Store.Name = p.StoreName
	return r
}

func ToPublicProductResponseList(list []domain.Product) []PublicProductResponse {
	out := make([]PublicProductResponse, 0, len(list))
	for i := range list {
		out = append(out, ToPublicProductResponse(&list[i]))
	}
	return out
}
