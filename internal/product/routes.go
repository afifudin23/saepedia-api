package product

import (
	"github.com/afifudin23/saepedia-api/internal/product/domain"
	"github.com/afifudin23/saepedia-api/internal/product/handler"
	"github.com/afifudin23/saepedia-api/internal/product/usecase"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.ProductRepository, storeRepo storedomain.StoreRepository) *Module {
	uc := usecase.New(repo, storeRepo)
	return &Module{handler: handler.New(uc)}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Public catalog — guest boleh lihat.
	pub := rg.Group("/products")
	pub.GET("", m.handler.ListPublic)
	pub.GET("/:id", m.handler.GetPublic)

	// Seller — CRUD produk sendiri (role aktif harus seller).
	seller := rg.Group("/seller/products")
	seller.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleSeller))
	{
		seller.GET("", m.handler.ListMine)
		seller.POST("", m.handler.Create)
		seller.PUT("/:id", m.handler.Update)
		seller.DELETE("/:id", m.handler.Delete)
	}
}
