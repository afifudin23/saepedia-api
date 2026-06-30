package admin

import (
	"github.com/afifudin23/saepedia-api/internal/admin/domain"
	"github.com/afifudin23/saepedia-api/internal/admin/handler"
	"github.com/afifudin23/saepedia-api/internal/admin/usecase"
	orderusecase "github.com/afifudin23/saepedia-api/internal/order/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.MonitorRepository, orderUC orderusecase.OrderUsecase, settings usecase.TimeSettingStore) *Module {
	uc := usecase.New(repo, orderUC, settings)
	return &Module{handler: handler.New(uc)}
}

// Semua endpoint admin butuh role aktif admin.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/admin")
	r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleAdmin))
	{
		r.GET("/summary", m.handler.Summary)
		r.GET("/users", m.handler.Users)
		r.GET("/stores", m.handler.Stores)
		r.GET("/products", m.handler.Products)
		r.GET("/orders", m.handler.Orders)
		r.GET("/deliveries", m.handler.Deliveries)
		r.GET("/overdue-orders", m.handler.OverdueOrders)

		r.GET("/simulate/now", m.handler.SimulateNow)
		r.POST("/simulate/advance-day", m.handler.SimulateAdvance)
		r.POST("/overdue/run", m.handler.RunOverdue)
	}
}
