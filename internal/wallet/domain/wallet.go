package domain

import (
	"context"
	"errors"
	"time"
)

var ErrInsufficientBalance = errors.New("insufficient wallet balance")

// Tipe transaksi dompet.
const (
	TxTopUp   = "topup"
	TxPayment = "payment"
	TxRefund  = "refund"
)

type Wallet struct {
	ID        string
	UserID    string
	Balance   int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WalletTransaction struct {
	ID           string
	WalletID     string
	Type         string
	Amount       int64
	BalanceAfter int64
	Description  string
	CreatedAt    time.Time
}

type WalletRepository interface {
	GetOrCreate(ctx context.Context, userID string) (*Wallet, error)
	// GetForUpdate memastikan wallet ada lalu mengunci barisnya (dipakai di dalam transaksi).
	GetForUpdate(ctx context.Context, userID string) (*Wallet, error)
	UpdateBalance(ctx context.Context, walletID string, newBalance int64) error
	AddTransaction(ctx context.Context, walletID, txType string, amount, balanceAfter int64, desc string) (*WalletTransaction, error)
	ListTransactions(ctx context.Context, userID string, limit, offset int) ([]WalletTransaction, int64, error)
	// BalanceByUserID mengembalikan saldo (0 bila belum punya wallet).
	BalanceByUserID(ctx context.Context, userID string) (int64, error)
}
