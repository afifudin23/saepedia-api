package handler

import (
	"net/http"

	"github.com/afifudin23/saepedia-api/internal/wallet/dto"
	"github.com/afifudin23/saepedia-api/internal/wallet/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.WalletUsecase
}

func New(uc usecase.WalletUsecase) *Handler {
	return &Handler{uc: uc}
}

// Get godoc
// @Summary   Saldo wallet (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /buyer/wallet [get]
func (h *Handler) Get(c *gin.Context) {
	wallet, err := h.uc.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToWalletResponse(wallet))
}

// TopUp godoc
// @Summary   Top-up dummy (buyer)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.TopUpRequest  true  "Jumlah top-up"
// @Success   200   {object}  response.Response
// @Router    /buyer/wallet/topup [post]
func (h *Handler) TopUp(c *gin.Context) {
	var req dto.TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	wallet, tx, err := h.uc.TopUp(c.Request.Context(), middleware.UserID(c), req.Amount)
	if err != nil {
		response.InternalServerError(c)
		return
	}

	response.Success(c, http.StatusOK, "top-up success", gin.H{
		"wallet":      dto.ToWalletResponse(wallet),
		"transaction": dto.ToTransactionResponse(tx),
	})
}

// History godoc
// @Summary   Riwayat transaksi wallet (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /buyer/wallet/transactions [get]
func (h *Handler) History(c *gin.Context) {
	p := pagination.Parse(c)
	txs, total, err := h.uc.History(c.Request.Context(), middleware.UserID(c), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToTransactionResponseList(txs))
}
