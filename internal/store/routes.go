package store

import (
	"github.com/afifudin23/saepedia-api/internal/store/domain"
	"github.com/afifudin23/saepedia-api/internal/store/handler"
	"github.com/afifudin23/saepedia-api/internal/store/usecase"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
	uc      usecase.StoreUsecase
}

func New(repo domain.StoreRepository) *Module {
	uc := usecase.New(repo)
	return &Module{handler: handler.New(uc), uc: uc}
}

// Usecase diekspos agar modul lain (product, order) bisa pakai logika toko.
func (m *Module) Usecase() usecase.StoreUsecase { return m.uc }

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Public — toko bisa dilihat guest.
	pub := rg.Group("/stores")
	pub.GET("", m.handler.ListPublic)
	pub.GET("/:id", m.handler.GetPublic)

	// Seller — kelola toko sendiri (role aktif harus seller).
	seller := rg.Group("/seller/store")
	seller.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleSeller))
	{
		seller.GET("", m.handler.GetMine)
		seller.POST("", m.handler.UpsertMine)
		seller.PUT("", m.handler.UpsertMine)
	}
}
