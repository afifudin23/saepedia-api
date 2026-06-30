package handler

import (
	"errors"
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/delivery/domain"
	"github.com/afifudin23/saepedia-api/internal/delivery/dto"
	"github.com/afifudin23/saepedia-api/internal/delivery/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.DeliveryUsecase
}

func New(uc usecase.DeliveryUsecase) *Handler {
	return &Handler{uc: uc}
}

// AvailableJobs — daftar job yang siap diambil (status Menunggu Pengirim).
// AvailableJobs godoc
// @Summary   Daftar job tersedia (driver)
// @Tags      Driver
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /driver/jobs [get]
func (h *Handler) AvailableJobs(c *gin.Context) {
	p := pagination.Parse(c)
	jobs, total, err := h.uc.AvailableJobs(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToJobResponseList(jobs))
}

// JobDetail godoc
// @Summary   Detail job tersedia (driver)
// @Tags      Driver
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   404  {object}  response.Response
// @Router    /driver/jobs/{id} [get]
func (h *Handler) JobDetail(c *gin.Context) {
	job, err := h.uc.JobDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToJobResponse(job))
}

// Take godoc
// @Summary   Ambil job → Sedang Dikirim (driver)
// @Tags      Driver
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   409  {object}  response.Response
// @Router    /driver/jobs/{id}/take [post]
func (h *Handler) Take(c *gin.Context) {
	job, err := h.uc.Take(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "job taken", dto.ToJobResponse(job))
}

// Complete godoc
// @Summary   Selesaikan job → Pesanan Selesai (driver)
// @Tags      Driver
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   422  {object}  response.Response
// @Router    /driver/jobs/{id}/complete [post]
func (h *Handler) Complete(c *gin.Context) {
	job, err := h.uc.Complete(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "job completed", dto.ToJobResponse(job))
}

// Dashboard godoc
// @Summary   Dashboard driver (job aktif, riwayat, earning)
// @Tags      Driver
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /driver/dashboard [get]
func (h *Handler) Dashboard(c *gin.Context) {
	dash, err := h.uc.Dashboard(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToDashboardResponse(dash))
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrJobNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrJobTaken):
		response.Conflict(c, err.Error())
	case errors.Is(err, domain.ErrJobNotYours):
		response.Forbidden(c, err.Error())
	case errors.Is(err, domain.ErrJobInvalidState):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
