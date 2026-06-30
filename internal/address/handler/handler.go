package handler

import (
	"errors"
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/address/domain"
	"github.com/afifudin23/saepedia-api/internal/address/dto"
	"github.com/afifudin23/saepedia-api/internal/address/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.AddressUsecase
}

func New(uc usecase.AddressUsecase) *Handler {
	return &Handler{uc: uc}
}

// Create godoc
// @Summary   Tambah alamat (buyer)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.UpsertAddressRequest  true  "Alamat"
// @Success   201   {object}  response.Response
// @Router    /buyer/addresses [post]
func (h *Handler) Create(c *gin.Context) {
	var req dto.UpsertAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	a, err := h.uc.Create(c.Request.Context(), middleware.UserID(c), toInput(req))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusCreated, "address created", dto.ToAddressResponse(a))
}

// Update godoc
// @Summary   Ubah alamat (buyer)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      string                    true  "Address ID"
// @Param     body  body      dto.UpsertAddressRequest  true  "Alamat"
// @Success   200   {object}  response.Response
// @Router    /buyer/addresses/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var req dto.UpsertAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	a, err := h.uc.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), toInput(req))
	if err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "address updated", dto.ToAddressResponse(a))
}

// Delete godoc
// @Summary   Hapus alamat (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Address ID"
// @Success   200  {object}  response.Response
// @Router    /buyer/addresses/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	err := h.uc.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "address deleted", nil)
}

// List godoc
// @Summary   Daftar alamat (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /buyer/addresses [get]
func (h *Handler) List(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.uc.List(c.Request.Context(), middleware.UserID(c), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToAddressResponseList(list))
}

func toInput(req dto.UpsertAddressRequest) usecase.AddressInput {
	return usecase.AddressInput{
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		FullAddress:   req.FullAddress,
		IsPrimary:     req.IsPrimary,
	}
}
