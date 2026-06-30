package handler

import (
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/admin/dto"
	"github.com/afifudin23/saepedia-api/internal/admin/usecase"
	orderdto "github.com/afifudin23/saepedia-api/internal/order/dto"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.AdminUsecase
}

func New(uc usecase.AdminUsecase) *Handler {
	return &Handler{uc: uc}
}

// Summary godoc
// @Summary   Ringkasan monitoring marketplace (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /admin/summary [get]
func (h *Handler) Summary(c *gin.Context) {
	s, err := h.uc.Summary(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToSummaryResponse(s))
}

// Users godoc
// @Summary   Monitoring users (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/users [get]
func (h *Handler) Users(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := h.uc.ListUsers(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToUserResponseList(rows))
}

// Stores godoc
// @Summary   Monitoring toko (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/stores [get]
func (h *Handler) Stores(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := h.uc.ListStores(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToStoreResponseList(rows))
}

// Products godoc
// @Summary   Monitoring produk (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/products [get]
func (h *Handler) Products(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := h.uc.ListProducts(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToProductResponseList(rows))
}

// Orders godoc
// @Summary   Monitoring order (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/orders [get]
func (h *Handler) Orders(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := h.uc.ListOrders(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToOrderRowResponseList(rows))
}

// Deliveries godoc
// @Summary   Monitoring pengiriman (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /admin/deliveries [get]
func (h *Handler) Deliveries(c *gin.Context) {
	p := pagination.Parse(c)
	rows, total, err := h.uc.ListDeliveries(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToDeliveryRowResponseList(rows))
}

// OverdueOrders godoc
// @Summary   Order yang sedang overdue (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /admin/overdue-orders [get]
func (h *Handler) OverdueOrders(c *gin.Context) {
	orders, err := h.uc.ListOverdue(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", orderdto.ToOrderResponseList(orders))
}

// SimulateNow godoc
// @Summary   Lihat waktu virtual saat ini (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /admin/simulate/now [get]
func (h *Handler) SimulateNow(c *gin.Context) {
	response.Success(c, http.StatusOK, "", gin.H{"now": h.uc.Now().Format("2006-01-02T15:04:05Z07:00")})
}

type advanceRequest struct {
	Days int `json:"days" binding:"required,gt=0,lte=365"`
}

// SimulateAdvance godoc
// @Summary   Majukan waktu N hari + jalankan overdue (admin)
// @Tags      Admin
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      object{days=int}  true  "Jumlah hari"
// @Success   200   {object}  response.Response
// @Router    /admin/simulate/advance-day [post]
func (h *Handler) SimulateAdvance(c *gin.Context) {
	var req advanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	res, err := h.uc.AdvanceDays(c.Request.Context(), req.Days)
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "time advanced and overdue orders processed", gin.H{
		"now":             res.Now.Format("2006-01-02T15:04:05Z07:00"),
		"offset_days":     res.OffsetDays,
		"processed_count": res.ProcessedCount,
		"processed":       orderdto.ToOrderResponseList(res.Processed),
	})
}

// RunOverdue godoc
// @Summary   Jalankan overdue handling manual (admin)
// @Tags      Admin
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /admin/overdue/run [post]
func (h *Handler) RunOverdue(c *gin.Context) {
	processed, err := h.uc.RunOverdue(c.Request.Context())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "overdue orders processed", gin.H{
		"processed_count": len(processed),
		"processed":       orderdto.ToOrderResponseList(processed),
	})
}
