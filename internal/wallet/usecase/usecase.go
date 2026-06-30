package usecase

import (
	"context"

	"github.com/afifudin23/saepedia-api/internal/wallet/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
)

type WalletUsecase interface {
	Get(ctx context.Context, userID string) (*domain.Wallet, error)
	TopUp(ctx context.Context, userID string, amount int64) (*domain.Wallet, *domain.WalletTransaction, error)
	History(ctx context.Context, userID string, limit, offset int) ([]domain.WalletTransaction, int64, error)
}

type walletUsecase struct {
	repo  domain.WalletRepository
	txMgr *tx.Manager
}

func New(repo domain.WalletRepository, txMgr *tx.Manager) WalletUsecase {
	return &walletUsecase{repo: repo, txMgr: txMgr}
}

func (uc *walletUsecase) Get(ctx context.Context, userID string) (*domain.Wallet, error) {
	return uc.repo.GetOrCreate(ctx, userID)
}

// TopUp: dummy top-up. Orkestrasi (kunci wallet → hitung saldo baru → simpan →
// catat transaksi) ada di usecase, dijalankan dalam satu transaksi.
func (uc *walletUsecase) TopUp(ctx context.Context, userID string, amount int64) (*domain.Wallet, *domain.WalletTransaction, error) {
	var wallet *domain.Wallet
	var record *domain.WalletTransaction

	err := uc.txMgr.Do(ctx, func(ctx context.Context) error {
		w, err := uc.repo.GetForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		newBalance := w.Balance + amount
		if err := uc.repo.UpdateBalance(ctx, w.ID, newBalance); err != nil {
			return err
		}
		w.Balance = newBalance
		wallet = w

		record, err = uc.repo.AddTransaction(ctx, w.ID, domain.TxTopUp, amount, newBalance, "dummy top-up")
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return wallet, record, nil
}

func (uc *walletUsecase) History(ctx context.Context, userID string, limit, offset int) ([]domain.WalletTransaction, int64, error) {
	return uc.repo.ListTransactions(ctx, userID, limit, offset)
}
