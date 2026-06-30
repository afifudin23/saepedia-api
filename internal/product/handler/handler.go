package handler

import (
	"errors"
	"net/http"

	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	"github.com/afifudin23/saepedia-api/internal/product/dto"
	"github.com/afifudin23/saepedia-api/internal/product/usecase"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.ProductUsecase
}

func New(uc usecase.ProductUsecase) *Handler {
	return &Handler{uc: uc}
}

// Create godoc
// @Summary   Tambah produk (seller)
// @Tags      Seller
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.CreateProductRequest  true  "Produk"
// @Success   201   {object}  response.Response
// @Router    /seller/products [post]
func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	product, err := h.uc.Create(c.Request.Context(), middleware.UserID(c), usecase.ProductInput{
		Name: req.Name, Description: req.Description, Price: req.Price, Stock: req.Stock, ImageURL: req.ImageURL,
	})
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "product created", dto.ToProductResponse(product))
}

// Update godoc
// @Summary   Ubah produk sendiri (seller)
// @Tags      Seller
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      string                    true  "Product ID"
// @Param     body  body      dto.UpdateProductRequest  true  "Produk"
// @Success   200   {object}  response.Response
// @Failure   403   {object}  response.Response
// @Router    /seller/products/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	product, err := h.uc.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), usecase.ProductInput{
		Name: req.Name, Description: req.Description, Price: req.Price, Stock: req.Stock, ImageURL: req.ImageURL,
	})
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "product updated", dto.ToProductResponse(product))
}

// Delete godoc
// @Summary   Hapus produk sendiri (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Product ID"
// @Success   200  {object}  response.Response
// @Failure   403  {object}  response.Response
// @Router    /seller/products/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.uc.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id")); err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "product deleted", nil)
}

// ListMine godoc
// @Summary   Daftar produk toko sendiri (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /seller/products [get]
func (h *Handler) ListMine(c *gin.Context) {
	p := pagination.Parse(c)
	products, total, err := h.uc.ListMine(c.Request.Context(), middleware.UserID(c), p.PerPage, p.Offset())
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToProductResponseList(products))
}

// ListPublic godoc
// @Summary  Katalog produk (publik)
// @Tags     Public
// @Produce  json
// @Param    search    query     string  false  "Cari nama produk"
// @Param    page      query     int     false  "Halaman"
// @Param    per_page  query     int     false  "Item per halaman"
// @Success  200       {object}  response.Response
// @Router   /products [get]
func (h *Handler) ListPublic(c *gin.Context) {
	p := pagination.Parse(c)
	products, total, err := h.uc.ListPublic(c.Request.Context(), c.Query("search"), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToPublicProductResponseList(products))
}

// GetPublic godoc
// @Summary  Detail produk (publik)
// @Tags     Public
// @Produce  json
// @Param    id   path      string  true  "Product ID"
// @Success  200  {object}  response.Response
// @Failure  404  {object}  response.Response
// @Router   /products/{id} [get]
func (h *Handler) GetPublic(c *gin.Context) {
	product, err := h.uc.GetPublic(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, productdomain.ErrProductNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToPublicProductResponse(product))
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, productdomain.ErrProductNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, productdomain.ErrNotProductOwner):
		response.Forbidden(c, err.Error())
	case errors.Is(err, storedomain.ErrStoreNotFound):
		response.BadRequest(c, "you must create a store before managing products")
	default:
		response.InternalServerError(c)
	}
}
