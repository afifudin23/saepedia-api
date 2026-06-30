package delivery

import (
	"github.com/afifudin23/saepedia-api/internal/delivery/domain"
	"github.com/afifudin23/saepedia-api/internal/delivery/handler"
	"github.com/afifudin23/saepedia-api/internal/delivery/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.DeliveryRepository, txMgr *tx.Manager) *Module {
	uc := usecase.New(repo, txMgr)
	return &Module{handler: handler.New(uc)}
}

// Semua endpoint driver butuh role aktif driver.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/driver")
	r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleDriver))
	{
		r.GET("/jobs", m.handler.AvailableJobs)
		r.GET("/jobs/:id", m.handler.JobDetail)
		r.POST("/jobs/:id/take", m.handler.Take)
		r.POST("/jobs/:id/complete", m.handler.Complete)
		r.GET("/dashboard", m.handler.Dashboard)
	}
}
