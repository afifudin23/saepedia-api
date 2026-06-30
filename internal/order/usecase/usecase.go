package usecase

import (
	"context"
	"errors"

	addressdomain "github.com/afifudin23/saepedia-api/internal/address/domain"
	cartdomain "github.com/afifudin23/saepedia-api/internal/cart/domain"
	cartusecase "github.com/afifudin23/saepedia-api/internal/cart/usecase"
	discountdomain "github.com/afifudin23/saepedia-api/internal/discount/domain"
	discountusecase "github.com/afifudin23/saepedia-api/internal/discount/usecase"
	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	productdomain "github.com/afifudin23/saepedia-api/internal/product/domain"
	storedomain "github.com/afifudin23/saepedia-api/internal/store/domain"
	walletdomain "github.com/afifudin23/saepedia-api/internal/wallet/domain"
	"github.com/afifudin23/saepedia-api/pkg/clock"
	"github.com/afifudin23/saepedia-api/pkg/tx"
)

// CheckoutSummary adalah rincian biaya sebelum/saat order dibuat.
type CheckoutSummary struct {
	Subtotal     int64
	Discount     int64
	DiscountCode string
	DiscountKind string
	DeliveryFee  int64
	Tax          int64
	Total        int64
}

type OrderUsecase interface {
	Preview(ctx context.Context, userID, deliveryMethod, discountCode string) (*CheckoutSummary, error)
	Checkout(ctx context.Context, userID, addressID, deliveryMethod, discountCode string) (*orderdomain.Order, error)
	GetForBuyer(ctx context.Context, userID, orderID string) (*orderdomain.Order, error)
	ListForBuyer(ctx context.Context, userID string, limit, offset int) ([]orderdomain.Order, int64, error)
	ListForSeller(ctx context.Context, userID string, limit, offset int) ([]orderdomain.Order, int64, error)
	GetForSeller(ctx context.Context, userID, orderID string) (*orderdomain.Order, error)
	ProcessOrder(ctx context.Context, userID, orderID string) (*orderdomain.Order, error)
	BuyerReport(ctx context.Context, userID string) (*orderdomain.BuyerReport, error)
	SellerReport(ctx context.Context, userID string) (*orderdomain.SellerReport, error)
	ListOverdue(ctx context.Context) ([]orderdomain.Order, error)
	RunOverdue(ctx context.Context) ([]orderdomain.Order, error)
}

type orderUsecase struct {
	repo        orderdomain.OrderRepository
	cartUC      cartusecase.CartUsecase
	addressRepo addressdomain.AddressRepository
	storeRepo   storedomain.StoreRepository
	discountUC  discountusecase.DiscountUsecase
	productRepo productdomain.ProductRepository
	walletRepo  walletdomain.WalletRepository
	txMgr       *tx.Manager
}

func New(
	repo orderdomain.OrderRepository,
	cartUC cartusecase.CartUsecase,
	addressRepo addressdomain.AddressRepository,
	storeRepo storedomain.StoreRepository,
	discountUC discountusecase.DiscountUsecase,
	productRepo productdomain.ProductRepository,
	walletRepo walletdomain.WalletRepository,
	txMgr *tx.Manager,
) OrderUsecase {
	return &orderUsecase{
		repo: repo, cartUC: cartUC, addressRepo: addressRepo, storeRepo: storeRepo,
		discountUC: discountUC, productRepo: productRepo, walletRepo: walletRepo, txMgr: txMgr,
	}
}

func (uc *orderUsecase) Preview(ctx context.Context, userID, deliveryMethod, discountCode string) (*CheckoutSummary, error) {
	fee, ok := orderdomain.DeliveryFee(deliveryMethod)
	if !ok {
		return nil, orderdomain.ErrInvalidDelivery
	}
	cart, err := uc.cartUC.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, orderdomain.ErrCartEmpty
	}

	var discount *discountusecase.ValidationResult
	if discountCode != "" {
		discount, err = uc.discountUC.Validate(ctx, discountCode, cart.Subtotal)
		if err != nil {
			return nil, err
		}
	}
	return computeSummary(cart.Subtotal, fee, discount), nil
}

// Checkout mengorkestrasi seluruh proses dalam SATU transaksi: kurangi stok,
// pakai voucher, buat order + riwayat, potong wallet, kosongkan cart.
func (uc *orderUsecase) Checkout(ctx context.Context, userID, addressID, deliveryMethod, discountCode string) (*orderdomain.Order, error) {
	fee, ok := orderdomain.DeliveryFee(deliveryMethod)
	if !ok {
		return nil, orderdomain.ErrInvalidDelivery
	}

	cart, err := uc.cartUC.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, orderdomain.ErrCartEmpty
	}

	address, err := uc.addressRepo.FindForUser(ctx, userID, addressID)
	if err != nil {
		if errors.Is(err, addressdomain.ErrAddressNotFound) {
			return nil, orderdomain.ErrAddressNotFound
		}
		return nil, err
	}

	// Validasi diskon (hanya SATU kode per checkout — voucher ATAU promo).
	var discount *discountusecase.ValidationResult
	if discountCode != "" {
		discount, err = uc.discountUC.Validate(ctx, discountCode, cart.Subtotal)
		if err != nil {
			return nil, err
		}
	}

	summary := computeSummary(cart.Subtotal, fee, discount)

	order := &orderdomain.Order{
		BuyerID:        userID,
		StoreID:        cart.StoreID,
		RecipientName:  address.RecipientName,
		Phone:          address.Phone,
		FullAddress:    address.FullAddress,
		DeliveryMethod: deliveryMethod,
		Subtotal:       summary.Subtotal,
		Discount:       summary.Discount,
		DiscountCode:   summary.DiscountCode,
		DeliveryFee:    summary.DeliveryFee,
		Tax:            summary.Tax,
		Total:          summary.Total,
		Status:         orderdomain.StatusDikemas, // status awal ditentukan usecase
		Items:          toOrderItems(cart.Items),
	}

	err = uc.txMgr.Do(ctx, func(ctx context.Context) error {
		// 1. Kurangi stok tiap item (cegah stok negatif).
		for _, it := range order.Items {
			if err := uc.productRepo.DecrementStock(ctx, it.ProductID, it.Quantity); err != nil {
				if errors.Is(err, productdomain.ErrInsufficientStock) {
					return orderdomain.ErrInsufficientStock
				}
				return err
			}
		}

		// 2. Pakai voucher (Promo tidak punya kuota).
		if discount != nil && discount.Kind == discountdomain.KindVoucher {
			if err := uc.discountUC.ConsumeVoucher(ctx, discount.ID); err != nil {
				if errors.Is(err, discountdomain.ErrVoucherRejected) {
					return orderdomain.ErrDiscountRejected
				}
				return err
			}
		}

		// 3. Buat order + riwayat status awal.
		if err := uc.repo.Create(ctx, order); err != nil {
			return err
		}
		if err := uc.repo.AddStatusHistory(ctx, order.ID, orderdomain.StatusDikemas, "order created after successful checkout"); err != nil {
			return err
		}

		// 4. Potong saldo wallet (cek kecukupan dulu).
		wallet, err := uc.walletRepo.GetForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		if wallet.Balance < order.Total {
			return orderdomain.ErrInsufficientFunds
		}
		newBalance := wallet.Balance - order.Total
		if err := uc.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
			return err
		}
		if _, err := uc.walletRepo.AddTransaction(ctx, wallet.ID, walletdomain.TxPayment, -order.Total, newBalance, "payment for order "+order.ID); err != nil {
			return err
		}

		// 5. Kosongkan cart.
		return uc.cartUC.Clear(ctx, userID)
	})
	if err != nil {
		return nil, err
	}

	return uc.repo.FindByIDForBuyer(ctx, userID, order.ID)
}

func (uc *orderUsecase) GetForBuyer(ctx context.Context, userID, orderID string) (*orderdomain.Order, error) {
	return uc.repo.FindByIDForBuyer(ctx, userID, orderID)
}

func (uc *orderUsecase) ListForBuyer(ctx context.Context, userID string, limit, offset int) ([]orderdomain.Order, int64, error) {
	return uc.repo.ListByBuyer(ctx, userID, limit, offset)
}

func (uc *orderUsecase) ListForSeller(ctx context.Context, userID string, limit, offset int) ([]orderdomain.Order, int64, error) {
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return uc.repo.ListByStore(ctx, store.ID, limit, offset)
}

func (uc *orderUsecase) GetForSeller(ctx context.Context, userID, orderID string) (*orderdomain.Order, error) {
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.repo.FindByIDForStore(ctx, store.ID, orderID)
}

// ProcessOrder: aturan transisi Sedang Dikemas → Menunggu Pengirim ada di usecase.
func (uc *orderUsecase) ProcessOrder(ctx context.Context, userID, orderID string) (*orderdomain.Order, error) {
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Pastikan order ada & milik toko ini.
	order, err := uc.repo.FindByIDForStore(ctx, store.ID, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != orderdomain.StatusDikemas {
		return nil, orderdomain.ErrInvalidTransition
	}

	err = uc.txMgr.Do(ctx, func(ctx context.Context) error {
		ok, err := uc.repo.UpdateStatusGuarded(ctx, orderID, orderdomain.StatusDikemas, orderdomain.StatusMenungguKirim)
		if err != nil {
			return err
		}
		if !ok {
			return orderdomain.ErrInvalidTransition
		}
		return uc.repo.AddStatusHistory(ctx, orderID, orderdomain.StatusMenungguKirim, "order processed by seller, ready for pickup")
	})
	if err != nil {
		return nil, err
	}
	return uc.repo.FindByIDForStore(ctx, store.ID, orderID)
}

func (uc *orderUsecase) BuyerReport(ctx context.Context, userID string) (*orderdomain.BuyerReport, error) {
	return uc.repo.BuyerReport(ctx, userID)
}

func (uc *orderUsecase) SellerReport(ctx context.Context, userID string) (*orderdomain.SellerReport, error) {
	store, err := uc.storeRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.repo.SellerReport(ctx, store.ID)
}

// ListOverdue: aturan SLA per metode (pakai waktu virtual) ada di usecase.
func (uc *orderUsecase) ListOverdue(ctx context.Context) ([]orderdomain.Order, error) {
	candidates, err := uc.repo.ListRefundCandidates(ctx)
	if err != nil {
		return nil, err
	}
	now := clock.Now()
	overdue := make([]orderdomain.Order, 0)
	for _, o := range candidates {
		if now.After(o.CreatedAt.Add(orderdomain.SLA(o.DeliveryMethod))) {
			overdue = append(overdue, o)
		}
	}
	return overdue, nil
}

// RunOverdue memproses tiap order overdue: refund + restore stok + Dikembalikan.
func (uc *orderUsecase) RunOverdue(ctx context.Context) ([]orderdomain.Order, error) {
	overdue, err := uc.ListOverdue(ctx)
	if err != nil {
		return nil, err
	}

	processed := make([]orderdomain.Order, 0, len(overdue))
	for i := range overdue {
		refunded, err := uc.refundOne(ctx, overdue[i].ID)
		if err != nil {
			if errors.Is(err, orderdomain.ErrInvalidTransition) {
				continue // sudah final/terlanjur diproses → idempotent skip
			}
			return processed, err
		}
		processed = append(processed, *refunded)
	}
	return processed, nil
}

func (uc *orderUsecase) refundOne(ctx context.Context, orderID string) (*orderdomain.Order, error) {
	var buyerID string
	err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
		order, err := uc.repo.GetForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		// Idempotensi: jangan refund order yang sudah final / sudah di-refund.
		if order.Refunded || order.Status == orderdomain.StatusDikembalikan || order.Status == orderdomain.StatusSelesai {
			return orderdomain.ErrInvalidTransition
		}
		buyerID = order.BuyerID

		// 1. Pulihkan stok.
		items, err := uc.repo.Items(ctx, orderID)
		if err != nil {
			return err
		}
		for _, it := range items {
			if err := uc.productRepo.RestoreStock(ctx, it.ProductID, it.Quantity); err != nil {
				return err
			}
		}

		// 2. Kembalikan dana ke wallet buyer.
		wallet, err := uc.walletRepo.GetForUpdate(ctx, order.BuyerID)
		if err != nil {
			return err
		}
		newBalance := wallet.Balance + order.Total
		if err := uc.walletRepo.UpdateBalance(ctx, wallet.ID, newBalance); err != nil {
			return err
		}
		if _, err := uc.walletRepo.AddTransaction(ctx, wallet.ID, walletdomain.TxRefund, order.Total, newBalance, "auto refund for overdue order "+orderID); err != nil {
			return err
		}

		// 3. Tandai Dikembalikan + riwayat.
		if err := uc.repo.MarkRefunded(ctx, orderID); err != nil {
			return err
		}
		return uc.repo.AddStatusHistory(ctx, orderID, orderdomain.StatusDikembalikan, "auto refund: order overdue (delivery SLA exceeded)")
	})
	if err != nil {
		return nil, err
	}
	return uc.repo.FindByIDForBuyer(ctx, buyerID, orderID)
}

// computeSummary: PPN 12% dari (subtotal - discount). Total = base + ongkir + PPN.
func computeSummary(subtotal, deliveryFee int64, discount *discountusecase.ValidationResult) *CheckoutSummary {
	s := &CheckoutSummary{Subtotal: subtotal, DeliveryFee: deliveryFee}
	if discount != nil {
		s.Discount = discount.Amount
		s.DiscountCode = discount.Code
		s.DiscountKind = discount.Kind
	}
	base := subtotal - s.Discount
	if base < 0 {
		base = 0
	}
	s.Tax = orderdomain.CalcTax(base)
	s.Total = base + deliveryFee + s.Tax
	return s
}

func toOrderItems(items []cartdomain.CartItem) []orderdomain.OrderItem {
	out := make([]orderdomain.OrderItem, 0, len(items))
	for _, it := range items {
		out = append(out, orderdomain.OrderItem{
			ProductID:   it.ProductID,
			ProductName: it.ProductName,
			Price:       it.Price,
			Quantity:    it.Quantity,
			Subtotal:    it.Subtotal,
		})
	}
	return out
}
