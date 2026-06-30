package review

import (
	"github.com/afifudin23/saepedia-api/internal/review/domain"
	"github.com/afifudin23/saepedia-api/internal/review/handler"
	"github.com/afifudin23/saepedia-api/internal/review/usecase"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(repo domain.ReviewRepository) *Module {
	uc := usecase.New(repo)
	return &Module{handler: handler.New(uc)}
}

// RegisterRoutes — semua endpoint review bersifat publik (boleh guest).
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/reviews")
	r.GET("", m.handler.List)
	r.POST("", m.handler.Create)
}
