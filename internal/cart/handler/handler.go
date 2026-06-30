package handler

import (
	"errors"
	"net/http"

	cartdomain "github.com/afifudin23/saepedia-api/internal/cart/domain"
	"github.com/afifudin23/saepedia-api/internal/cart/dto"
	"github.com/afifudin23/saepedia-api/internal/cart/usecase"
	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.CartUsecase
}

func New(uc usecase.CartUsecase) *Handler {
	return &Handler{uc: uc}
}

// Get godoc
// @Summary   Lihat cart (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /buyer/cart [get]
func (h *Handler) Get(c *gin.Context) {
	cart, err := h.uc.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToCartResponse(cart))
}

// AddItem godoc
// @Summary   Tambah produk ke cart (single-store)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.AddItemRequest  true  "Item"
// @Success   200   {object}  response.Response
// @Failure   409   {object}  response.Response
// @Router    /buyer/cart/items [post]
func (h *Handler) AddItem(c *gin.Context) {
	var req dto.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	cart, err := h.uc.AddItem(c.Request.Context(), middleware.UserID(c), req.ProductID, req.Quantity)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "item added to cart", dto.ToCartResponse(cart))
}

// UpdateItem godoc
// @Summary   Ubah qty item cart
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     productID  path      string                 true  "Product ID"
// @Param     body       body      dto.UpdateItemRequest  true  "Qty baru"
// @Success   200        {object}  response.Response
// @Router    /buyer/cart/items/{productID} [put]
func (h *Handler) UpdateItem(c *gin.Context) {
	var req dto.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	cart, err := h.uc.UpdateItem(c.Request.Context(), middleware.UserID(c), c.Param("productID"), req.Quantity)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "cart item updated", dto.ToCartResponse(cart))
}

// RemoveItem godoc
// @Summary   Hapus item dari cart
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Param     productID  path      string  true  "Product ID"
// @Success   200        {object}  response.Response
// @Router    /buyer/cart/items/{productID} [delete]
func (h *Handler) RemoveItem(c *gin.Context) {
	cart, err := h.uc.RemoveItem(c.Request.Context(), middleware.UserID(c), c.Param("productID"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "cart item removed", dto.ToCartResponse(cart))
}

// Clear godoc
// @Summary   Kosongkan cart
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /buyer/cart [delete]
func (h *Handler) Clear(c *gin.Context) {
	if err := h.uc.Clear(c.Request.Context(), middleware.UserID(c)); err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "cart cleared", nil)
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, cartdomain.ErrDifferentStore):
		response.Conflict(c, err.Error())
	case errors.Is(err, cartdomain.ErrItemNotInCart):
		response.NotFound(c, err.Error())
	case errors.Is(err, productdomain.ErrProductNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, productdomain.ErrInsufficientStock):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
