package handler

import (
	"errors"
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/discount/domain"
	"github.com/afifudin23/saepedia-api/internal/discount/dto"
	"github.com/afifudin23/saepedia-api/internal/discount/usecase"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.DiscountUsecase
}

func New(uc usecase.DiscountUsecase) *Handler {
	return &Handler{uc: uc}
}

// generate dipakai oleh GenerateVoucher & GeneratePromo.
func (h *Handler) generate(c *gin.Context, kind string) {
	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	expires, err := req.ParseExpiresAt()
	if err != nil {
		response.BadRequest(c, "expires_at must be RFC3339, e.g. 2026-12-31T23:59:59Z")
		return
	}

	d, err := h.uc.Generate(c.Request.Context(), kind, usecase.GenerateInput{
		Code:          req.Code,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		MaxDiscount:   req.MaxDiscount,
		MinSpend:      req.MinSpend,
		ExpiresAt:     expires,
		UsageLimit:    req.UsageLimit,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCodeExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusCreated, kind+" created", dto.ToDiscountResponse(d))
}

// GenerateVoucher godoc
// @Summary   Generate voucher (admin)
// @Tags      Admin
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.GenerateRequest  true  "Data voucher"
// @Success   201   {object}  response.Response
// @Failure   409   {object}  response.Response
// @Router    /admin/vouchers [post]
func (h *Handler) GenerateVoucher(c *gin.Context) { h.generate(c, domain.KindVoucher) }

// GeneratePromo godoc
// @Summary   Generate promo (admin)
// @Tags      Admin
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.GenerateRequest  true  "Data promo (usage_limit diabaikan)"
// @Success   201   {object}  response.Response
// @Failure   409   {object}  response.Response
// @Router    /admin/promos [post]
func (h *Handler) GeneratePromo(c *gin.Context) { h.generate(c, domain.KindPromo) }

func (h *Handler) list(c *gin.Context, kind string) {
	p := pagination.Parse(c)
	items, total, err := h.uc.List(c.Request.Context(), kind, p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToDiscountResponseList(items))
}

// ListVouchers godoc
// @Summary   Daftar voucher (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/vouchers [get]
func (h *Handler) ListVouchers(c *gin.Context) { h.list(c, domain.KindVoucher) }

// ListPromos godoc
// @Summary   Daftar promo (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/promos [get]
func (h *Handler) ListPromos(c *gin.Context) { h.list(c, domain.KindPromo) }

// Detail godoc
// @Summary   Detail voucher/promo by ID (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Discount ID"
// @Success   200  {object}  response.Response
// @Failure   404  {object}  response.Response
// @Router    /admin/vouchers/{id} [get]
// @Router    /admin/promos/{id} [get]
func (h *Handler) Detail(c *gin.Context) {
	d, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrDiscountNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToDiscountResponse(d))
}
