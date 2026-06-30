// internal/router/router.go
package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/afifudin23/saepedia-api/config"
	"github.com/afifudin23/saepedia-api/internal/address"
	addressrepo "github.com/afifudin23/saepedia-api/internal/address/repository"
	"github.com/afifudin23/saepedia-api/internal/admin"
	adminrepo "github.com/afifudin23/saepedia-api/internal/admin/repository"
	"github.com/afifudin23/saepedia-api/internal/auth"
	authrepo "github.com/afifudin23/saepedia-api/internal/auth/repository"
	"github.com/afifudin23/saepedia-api/internal/cart"
	cartrepo "github.com/afifudin23/saepedia-api/internal/cart/repository"
	"github.com/afifudin23/saepedia-api/internal/delivery"
	deliveryrepo "github.com/afifudin23/saepedia-api/internal/delivery/repository"
	"github.com/afifudin23/saepedia-api/internal/discount"
	discountrepo "github.com/afifudin23/saepedia-api/internal/discount/repository"
	"github.com/afifudin23/saepedia-api/internal/order"
	orderrepo "github.com/afifudin23/saepedia-api/internal/order/repository"
	"github.com/afifudin23/saepedia-api/internal/product"
	productrepo "github.com/afifudin23/saepedia-api/internal/product/repository"
	"github.com/afifudin23/saepedia-api/internal/review"
	reviewrepo "github.com/afifudin23/saepedia-api/internal/review/repository"
	settingrepo "github.com/afifudin23/saepedia-api/internal/setting/repository"
	"github.com/afifudin23/saepedia-api/internal/store"
	storerepo "github.com/afifudin23/saepedia-api/internal/store/repository"
	userrepo "github.com/afifudin23/saepedia-api/internal/user/repository"
	"github.com/afifudin23/saepedia-api/internal/wallet"
	walletrepo "github.com/afifudin23/saepedia-api/internal/wallet/repository"
	"github.com/afifudin23/saepedia-api/pkg/clock"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/response"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Setup merakit semua dependency & route, lalu mengembalikan gin engine.
func Setup(db *gorm.DB, accessKey string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	// Swagger UI per versi: http://localhost:<port>/docs/v1/index.html
	docsBase := "/docs/" + config.APIVersion
	docsRedirect := func(c *gin.Context) { c.Redirect(http.StatusFound, docsBase+"/index.html") }
	docsHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/docs", docsRedirect)   // /docs            → /docs/v1/index.html
	r.GET(docsBase, docsRedirect)  // /docs/v1         → /docs/v1/index.html
	r.GET(docsBase+"/*any", func(c *gin.Context) {
		// /docs/v1/ (root, trailing slash) ikut diarahkan ke index.html.
		if a := c.Param("any"); a == "/" || a == "" {
			docsRedirect(c)
			return
		}
		docsHandler(c)
	})

	r.GET("/", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Welcome to SEAPEDIA API", gin.H{
			"app":      config.AppConfig.AppName,
			"version":  config.Version,
			"docs":     docsBase + "/index.html",
			"health":   "/ping",
			"base_url": config.APIBasePath(),
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "pong", gin.H{
			"app":     config.AppConfig.AppName,
			"version": config.Version,
		})
	})

	// ── Repositories (satu sumber kebenaran per tabel) ─────────────
	userRepo := userrepo.New(db)
	reviewRepo := reviewrepo.New(db)
	storeRepo := storerepo.New(db)
	productRepo := productrepo.New(db)
	walletRepo := walletrepo.New(db)
	addressRepo := addressrepo.New(db)
	cartRepo := cartrepo.New(db)
	orderRepo := orderrepo.New(db)
	discountRepo := discountrepo.New(db)
	deliveryRepo := deliveryrepo.New(db)
	adminRepo := adminrepo.New(db)
	settingRepo := settingrepo.New(db)
	revocationRepo := authrepo.NewRevocation(db)

	// Transaction manager: dipakai usecase untuk membungkus operasi multi-repo.
	txMgr := tx.NewManager(db)

	// Muat offset simulasi waktu yang tersimpan ke clock global.
	if offset, err := settingRepo.GetTimeOffset(context.Background()); err == nil {
		clock.SetOffset(offset)
	}

	// Pasang pengecek denylist token (logout) ke middleware Auth.
	middleware.SetRevocationChecker(func(jti string) bool {
		revoked, _ := revocationRepo.IsRevoked(context.Background(), jti)
		return revoked
	})

	// ── Modul fitur ────────────────────────────────────────────────
	authModule := auth.New(userRepo, walletRepo, revocationRepo, accessKey) // walletRepo = WalletReader
	reviewModule := review.New(reviewRepo)
	storeModule := store.New(storeRepo)
	productModule := product.New(productRepo, storeRepo)
	walletModule := wallet.New(walletRepo, txMgr)
	addressModule := address.New(addressRepo, txMgr)
	cartModule := cart.New(cartRepo, productRepo)
	discountModule := discount.New(discountRepo)
	deliveryModule := delivery.New(deliveryRepo, txMgr)
	orderModule := order.New(orderRepo, cartModule.Usecase(), addressRepo, storeRepo, discountModule.Usecase(), productRepo, walletRepo, txMgr)
	adminModule := admin.New(adminRepo, orderModule.Usecase(), settingRepo)

	// ── Routes di bawah /api/<versi> (lihat config.APIVersion) ─────
	api := r.Group(config.APIBasePath())
	authModule.RegisterRoutes(api)
	reviewModule.RegisterRoutes(api)
	storeModule.RegisterRoutes(api)
	productModule.RegisterRoutes(api)
	walletModule.RegisterRoutes(api)
	addressModule.RegisterRoutes(api)
	cartModule.RegisterRoutes(api)
	discountModule.RegisterRoutes(api)
	deliveryModule.RegisterRoutes(api)
	orderModule.RegisterRoutes(api)
	adminModule.RegisterRoutes(api)

	return r
}
