package handler

import (
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/review/dto"
	"github.com/afifudin23/saepedia-api/internal/review/usecase"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.ReviewUsecase
}

func New(uc usecase.ReviewUsecase) *Handler {
	return &Handler{uc: uc}
}

// Create godoc
// @Summary  Kirim review aplikasi (publik, guest boleh)
// @Tags     Reviews
// @Accept   json
// @Produce  json
// @Param    body  body      dto.CreateReviewRequest  true  "Review"
// @Success  201   {object}  response.Response
// @Failure  400   {object}  response.Response
// @Router   /reviews [post]
func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	review, err := h.uc.Create(c.Request.Context(), req.ReviewerName, req.Rating, req.Comment)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Success(c, http.StatusCreated, "review submitted", dto.ToReviewResponse(review))
}

// List godoc
// @Summary  Daftar review aplikasi (publik)
// @Tags     Reviews
// @Produce  json
// @Param    page      query     int  false  "Halaman"
// @Param    per_page  query     int  false  "Item per halaman"
// @Success  200       {object}  response.Response
// @Router   /reviews [get]
func (h *Handler) List(c *gin.Context) {
	p := pagination.Parse(c)
	reviews, total, err := h.uc.List(c.Request.Context(), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToReviewResponseList(reviews))
}
