package dto

import (
	"time"

	"github.com/afifudin23/saepedia-api/internal/wallet/domain"
)

type TopUpRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

type WalletResponse struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

type TransactionResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Amount       int64  `json:"amount"`
	BalanceAfter int64  `json:"balance_after"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}

func ToWalletResponse(w *domain.Wallet) WalletResponse {
	return WalletResponse{ID: w.ID, Balance: w.Balance}
}

func ToTransactionResponse(t *domain.WalletTransaction) TransactionResponse {
	return TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		BalanceAfter: t.BalanceAfter,
		Description:  t.Description,
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
	}
}

func ToTransactionResponseList(list []domain.WalletTransaction) []TransactionResponse {
	out := make([]TransactionResponse, 0, len(list))
	for i := range list {
		out = append(out, ToTransactionResponse(&list[i]))
	}
	return out
}
