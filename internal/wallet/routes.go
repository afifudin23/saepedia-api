package wallet

import (
	"github.com/afifudin23/saepedia-api/internal/wallet/domain"
	"github.com/afifudin23/saepedia-api/internal/wallet/handler"
	"github.com/afifudin23/saepedia-api/internal/wallet/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.WalletRepository, txMgr *tx.Manager) *Module {
	uc := usecase.New(repo, txMgr)
	return &Module{handler: handler.New(uc)}
}

// Wallet hanya untuk role aktif buyer.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/buyer/wallet")
	r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleBuyer))
	{
		r.GET("", m.handler.Get)
		r.POST("/topup", m.handler.TopUp)
		r.GET("/transactions", m.handler.History)
	}
}
