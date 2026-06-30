package cart

import (
	"github.com/afifudin23/saepedia-api/internal/cart/domain"
	"github.com/afifudin23/saepedia-api/internal/cart/handler"
	"github.com/afifudin23/saepedia-api/internal/cart/usecase"
	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
	uc      usecase.CartUsecase
}

func New(repo domain.CartRepository, productRepo productdomain.ProductRepository) *Module {
	uc := usecase.New(repo, productRepo)
	return &Module{handler: handler.New(uc), uc: uc}
}

// Usecase diekspos agar modul order bisa membaca & mengosongkan cart.
func (m *Module) Usecase() usecase.CartUsecase { return m.uc }

// Cart hanya untuk role aktif buyer.
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/buyer/cart")
	r.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleBuyer))
	{
		r.GET("", m.handler.Get)
		r.POST("/items", m.handler.AddItem)
		r.PUT("/items/:productID", m.handler.UpdateItem)
		r.DELETE("/items/:productID", m.handler.RemoveItem)
		r.DELETE("", m.handler.Clear)
	}
}
