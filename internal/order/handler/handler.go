package handler

import (
	"errors"
	"net/http"

	discountdomain "github.com/afifudin23/saepedia-api/internal/discount/domain"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/internal/order/dto"
	"github.com/afifudin23/saepedia-api/internal/order/usecase"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/pagination"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	uc usecase.OrderUsecase
}

func New(uc usecase.OrderUsecase) *Handler {
	return &Handler{uc: uc}
}

// Preview — hitung ringkasan checkout (subtotal, diskon, ongkir, PPN, total).
// Preview godoc
// @Summary   Hitung ringkasan checkout (buyer)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.PreviewRequest  true  "Metode kirim + diskon opsional"
// @Success   200   {object}  response.Response
// @Router    /buyer/checkout/preview [post]
func (h *Handler) Preview(c *gin.Context) {
	var req dto.PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	summary, err := h.uc.Preview(c.Request.Context(), middleware.UserID(c), req.DeliveryMethod, req.DiscountCode)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToSummaryResponse(summary))
}

// Checkout — buat order dari cart.
// Checkout godoc
// @Summary   Checkout dari cart (buyer)
// @Tags      Buyer
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      dto.CheckoutRequest  true  "Alamat + metode kirim + diskon opsional"
// @Success   201   {object}  response.Response
// @Failure   422   {object}  response.Response
// @Router    /buyer/checkout [post]
func (h *Handler) Checkout(c *gin.Context) {
	var req dto.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	order, err := h.uc.Checkout(c.Request.Context(), middleware.UserID(c), req.AddressID, req.DeliveryMethod, req.DiscountCode)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "checkout success", dto.ToOrderResponse(order))
}

// BuyerList — riwayat order buyer.
// BuyerList godoc
// @Summary   Riwayat order (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /buyer/orders [get]
func (h *Handler) BuyerList(c *gin.Context) {
	p := pagination.Parse(c)
	orders, total, err := h.uc.ListForBuyer(c.Request.Context(), middleware.UserID(c), p.PerPage, p.Offset())
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToOrderResponseList(orders))
}

// BuyerDetail — detail order buyer (termasuk timeline status).
// BuyerDetail godoc
// @Summary   Detail order + timeline status (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   404  {object}  response.Response
// @Router    /buyer/orders/{id} [get]
func (h *Handler) BuyerDetail(c *gin.Context) {
	order, err := h.uc.GetForBuyer(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToOrderResponse(order))
}

// BuyerReport — ringkasan pengeluaran buyer.
// BuyerReport godoc
// @Summary   Ringkasan pengeluaran (buyer)
// @Tags      Buyer
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /buyer/reports [get]
func (h *Handler) BuyerReport(c *gin.Context) {
	report, err := h.uc.BuyerReport(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToBuyerReportResponse(report))
}

// SellerIncoming — daftar order masuk untuk toko seller.
// SellerIncoming godoc
// @Summary   Daftar order masuk (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Param     page      query     int  false  "Halaman"
// @Param     per_page  query     int  false  "Item per halaman"
// @Success   200       {object}  response.Response
// @Router    /seller/orders [get]
func (h *Handler) SellerIncoming(c *gin.Context) {
	p := pagination.Parse(c)
	orders, total, err := h.uc.ListForSeller(c.Request.Context(), middleware.UserID(c), p.PerPage, p.Offset())
	if err != nil {
		if errors.Is(err, storedomain.ErrStoreNotFound) {
			response.BadRequest(c, "you must create a store first")
			return
		}
		response.InternalServerError(c)
		return
	}
	response.List(c, p.Page, p.PerPage, int(total), "", dto.ToOrderResponseList(orders))
}

// SellerDetail — detail satu order milik toko seller.
// SellerDetail godoc
// @Summary   Detail order toko (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   404  {object}  response.Response
// @Router    /seller/orders/{id} [get]
func (h *Handler) SellerDetail(c *gin.Context) {
	order, err := h.uc.GetForSeller(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToOrderResponse(order))
}

// SellerProcess — Sedang Dikemas → Menunggu Pengirim.
// SellerProcess godoc
// @Summary   Proses order: Sedang Dikemas → Menunggu Pengirim (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      string  true  "Order ID"
// @Success   200  {object}  response.Response
// @Failure   422  {object}  response.Response
// @Router    /seller/orders/{id}/process [post]
func (h *Handler) SellerProcess(c *gin.Context) {
	order, err := h.uc.ProcessOrder(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "order processed", dto.ToOrderResponse(order))
}

// SellerReport — ringkasan pendapatan seller.
// SellerReport godoc
// @Summary   Ringkasan pendapatan (seller)
// @Tags      Seller
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  response.Response
// @Router    /seller/reports [get]
func (h *Handler) SellerReport(c *gin.Context) {
	report, err := h.uc.SellerReport(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		if errors.Is(err, storedomain.ErrStoreNotFound) {
			response.BadRequest(c, "you must create a store first")
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "", dto.ToSellerReportResponse(report))
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, orderdomain.ErrCartEmpty):
		response.UnprocessableEntity(c, err.Error())
	case errors.Is(err, orderdomain.ErrInvalidDelivery):
		response.BadRequest(c, err.Error())
	case errors.Is(err, orderdomain.ErrAddressNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, orderdomain.ErrInsufficientStock):
		response.UnprocessableEntity(c, err.Error())
	case errors.Is(err, orderdomain.ErrInsufficientFunds):
		response.UnprocessableEntity(c, err.Error())
	case errors.Is(err, orderdomain.ErrOrderNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, orderdomain.ErrInvalidTransition):
		response.UnprocessableEntity(c, err.Error())
	case errors.Is(err, orderdomain.ErrDiscountRejected):
		response.UnprocessableEntity(c, err.Error())
	// Error validasi diskon dari modul discount.
	case errors.Is(err, discountdomain.ErrDiscountNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, discountdomain.ErrDiscountExpired),
		errors.Is(err, discountdomain.ErrDiscountUsedUp),
		errors.Is(err, discountdomain.ErrMinSpendNotMet):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.InternalServerError(c)
	}
}
