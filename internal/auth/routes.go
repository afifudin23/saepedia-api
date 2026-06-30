package auth

import (
	"github.com/afifudin23/saepedia-api/internal/auth/domain"
	"github.com/afifudin23/saepedia-api/internal/auth/handler"
	"github.com/afifudin23/saepedia-api/internal/auth/usecase"
	userdomain "github.com/afifudin23/saepedia-api/internal/user/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
}

func New(userRepo userdomain.UserRepository, wallet domain.WalletReader, revoker domain.TokenRevoker, accessKey string) *Module {
	uc := usecase.New(userRepo, wallet, revoker, accessKey)
	return &Module{handler: handler.New(uc)}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	r := rg.Group("/auth")

	// Public
	r.POST("/register", m.handler.Register)
	r.POST("/login", m.handler.Login)

	// Authenticated (token wajib, role aktif boleh kosong)
	authed := r.Group("")
	authed.Use(middleware.Auth())
	{
		authed.POST("/logout", m.handler.Logout)
		authed.POST("/select-role", m.handler.SelectRole)
		authed.GET("/me", m.handler.Me)
		authed.GET("/balance-summary", m.handler.BalanceSummary)
	}
}
