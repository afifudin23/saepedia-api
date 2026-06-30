package handler

import (
	"errors"
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/store/domain"
	"github.com/afifudin23/saepedia-api/internal/store/dto"
	"github.com/afifudin23/saepedia-api/internal/store/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.StoreUsecase
}

func New(uc usecase.StoreUsecase) *Handler {
	return &Handler{uc: uc}
}

// UpsertMine godoc
// @Summary   Buat/ubah toko sendiri (seller)
// @Tags      Seller
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.UpsertStoreRequest  true  "Data toko"
// @Success   200   {object}  response.Response
// @Failure   409   {object}  response.Response
// @Router    /seller/store [post]
func (h *Handler) UpsertMine(c *gin.Context) {
	var req dto.UpsertStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	store, err := h.uc.CreateOrUpdate(c.Request.Context(), middleware.UserID(c), req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrStoreNameExists):
			response.Conflict(c, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.Success(c, http.StatusOK, "store saved", dto.ToStoreResponse(store))
}

// GetMine godoc
// @Summary   Detail toko sendiri (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Failure   404  {object}  response.Response
// @Router    /seller/store [get]
func (h *Handler) GetMine(c *gin.Context) {
	store, err := h.uc.GetMine(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			response.NotFound(c, "you don't have a store yet")
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToStoreResponse(store))
}

// GetPublic godoc
// @Summary  Detail toko (publik)
// @Tags     Public
// @Produce  json
// @Param    id   path      string  true  "Store ID"
// @Success  200  {object}  response.Response
// @Failure  404  {object}  response.Response
// @Router   /stores/{id} [get]
func (h *Handler) GetPublic(c *gin.Context) {
	store, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrStoreNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToPublicStoreSummary(store))
}

// ListPublic godoc
// @Summary  Daftar toko (publik)
// @Tags     Public
// @Produce  json
// @Param    page      query     int  false  "Halaman"
// @Param    per_page  query     int  false  "Item per halaman"
// @Success  200       {object}  response.Response
// @Router   /stores [get]
func (h *Handler) ListPublic(c *gin.Context) {
	p := pagination.Parse(c)
	stores, total, err := h.uc.List(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToPublicStoreSummaryList(stores))
}
