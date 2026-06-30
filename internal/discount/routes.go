package discount

import (
	"github.com/afifudin23/saepedia-api/internal/discount/domain"
	"github.com/afifudin23/saepedia-api/internal/discount/handler"
	"github.com/afifudin23/saepedia-api/internal/discount/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
	uc      usecase.DiscountUsecase
}

func New(repo domain.DiscountRepository) *Module {
	uc := usecase.New(repo)
	return &Module{handler: handler.New(uc), uc: uc}
}

// Usecase diekspos agar modul order bisa memvalidasi kode diskon saat checkout.
func (m *Module) Usecase() usecase.DiscountUsecase { return m.uc }

// Admin-only: generate, list, dan detail voucher & promo (Level 4 + Level 6 UI).
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	admin.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleAdmin))
	{
		admin.POST("/vouchers", m.handler.GenerateVoucher)
		admin.GET("/vouchers", m.handler.ListVouchers)
		admin.GET("/vouchers/:id", m.handler.Detail)

		admin.POST("/promos", m.handler.GeneratePromo)
		admin.GET("/promos", m.handler.ListPromos)
		admin.GET("/promos/:id", m.handler.Detail)
	}
}
