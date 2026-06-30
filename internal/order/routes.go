package order

import (
	addressdomain "github.com/afifudin23/saepedia-api/internal/address/domain"
	cartusecase "github.com/afifudin23/saepedia-api/internal/cart/usecase"
	discountusecase "github.com/afifudin23/saepedia-api/internal/discount/usecase"
	"github.com/afifudin23/saepedia-api/internal/order/domain"
	"github.com/afifudin23/saepedia-api/internal/order/handler"
	"github.com/afifudin23/saepedia-api/internal/order/usecase"
	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
	walletdomain "github.com/afifudin23/saepedia-api/internal/wallet/domain"
	"github.com/afifudin23/saepedia-api/pkg/middleware"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"github.com/gin-gonic/gin"
)

type Module struct {
	handler *handler.Handler
	uc      usecase.OrderUsecase
}

func New(
	repo domain.OrderRepository,
	cartUC cartusecase.CartUsecase,
	addressRepo addressdomain.AddressRepository,
	storeRepo storedomain.StoreRepository,
	discountUC discountusecase.DiscountUsecase,
	productRepo productdomain.ProductRepository,
	walletRepo walletdomain.WalletRepository,
	txMgr *tx.Manager,
) *Module {
	uc := usecase.New(repo, cartUC, addressRepo, storeRepo, discountUC, productRepo, walletRepo, txMgr)
	return &Module{handler: handler.New(uc), uc: uc}
}

// Usecase diekspos agar modul admin bisa menjalankan overdue handling.
func (m *Module) Usecase() usecase.OrderUsecase { return m.uc }

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	// Buyer — checkout, riwayat order, laporan pengeluaran.
	buyer := rg.Group("/buyer")
	buyer.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleBuyer))
	{
		buyer.POST("/checkout/preview", m.handler.Preview)
		buyer.POST("/checkout", m.handler.Checkout)
		buyer.GET("/orders", m.handler.BuyerList)
		buyer.GET("/orders/:id", m.handler.BuyerDetail)
		buyer.GET("/reports", m.handler.BuyerReport)
	}

	// Seller — order masuk, proses order, laporan pendapatan.
	seller := rg.Group("/seller")
	seller.Use(middleware.Auth(), middleware.RequireRole(middleware.RoleSeller))
	{
		seller.GET("/orders", m.handler.SellerIncoming)
		seller.GET("/orders/:id", m.handler.SellerDetail)
		seller.POST("/orders/:id/process", m.handler.SellerProcess)
		seller.GET("/reports", m.handler.SellerReport)
	}
}
