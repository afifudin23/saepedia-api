package address

import (
	"github.com/afifudin23/saepedia-api/internal/address/domain"
	"github.com/afifudin23/saepedia-api/internal/address/handler"
	"github.com/afifudin23/saepedia-api/internal/address/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.AddressRepository, txMgr *tx.Manager) *Module {
	uc := usecase.New(repo, txMgr)
	return &Module{handler: handler.New(uc)}
}

// Alamat pengiriman hanya untuk role aktif buyer.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/buyer/addresses")
	r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleBuyer))
	{
		r.GET("", m.handler.List)
		r.POST("", m.handler.Create)
		r.PUT("/:id", m.handler.Update)
		r.DELETE("/:id", m.handler.Delete)
	}
}
