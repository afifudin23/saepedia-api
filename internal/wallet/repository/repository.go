package repository

import (
	"context"
	"errors"
	"time"

	"github.com/afifudin23/saepedia-api/internal/wallet/domain"
	"github.com/afifudin23/saepedia-api/pkg/tx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   int64  `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (WalletModel) TableName() string { return "wallets" }

type WalletTransactionModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	WalletID     string `gorm:"type:uuid;not null;index"`
	Type         string `gorm:"not null"`
	Amount       int64  `gorm:"not null"`
	BalanceAfter int64  `gorm:"not null"`
	Description  string `gorm:"not null;default:''"`
	CreatedAt    time.Time
}

func (WalletTransactionModel) TableName() string { return "wallet_transactions" }

func (m WalletModel) toDomain() *domain.Wallet {
	return &domain.Wallet{
		ID: m.ID, UserID: m.UserID, Balance: m.Balance,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (m WalletTransactionModel) toDomain() domain.WalletTransaction {
	return domain.WalletTransaction{
		ID: m.ID, WalletID: m.WalletID, Type: m.Type,
		Amount: m.Amount, BalanceAfter: m.BalanceAfter,
		Description: m.Description, CreatedAt: m.CreatedAt,
	}
}

type walletRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetOrCreate(ctx context.Context, userID string) (*domain.Wallet, error) {
	model := WalletModel{UserID: userID}
	// FirstOrCreate aman dari race lewat unique index user_id.
	err := tx.DB(ctx, r.db).
		Where(WalletModel{UserID: userID}).
		FirstOrCreate(&model).Error
	if err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

// GetForUpdate memastikan wallet ada lalu mengunci barisnya. WAJIB dipanggil di
// dalam transaksi (lewat tx.Manager) agar locking berarti.
func (r *walletRepository) GetForUpdate(ctx context.Context, userID string) (*domain.Wallet, error) {
	db := tx.DB(ctx, r.db)
	var wallet WalletModel
	if err := db.Where(WalletModel{UserID: userID}).FirstOrCreate(&wallet).Error; err != nil {
		return nil, err
	}
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		return nil, err
	}
	return wallet.toDomain(), nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID string, newBalance int64) error {
	return tx.DB(ctx, r.db).Model(&WalletModel{}).
		Where("id = ?", walletID).
		Updates(map[string]any{"balance": newBalance, "updated_at": time.Now()}).Error
}

func (r *walletRepository) AddTransaction(ctx context.Context, walletID, txType string, amount, balanceAfter int64, desc string) (*domain.WalletTransaction, error) {
	model := WalletTransactionModel{
		WalletID:     walletID,
		Type:         txType,
		Amount:       amount,
		BalanceAfter: balanceAfter,
		Description:  desc,
	}
	if err := tx.DB(ctx, r.db).Create(&model).Error; err != nil {
		return nil, err
	}
	t := model.toDomain()
	return &t, nil
}

func (r *walletRepository) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]domain.WalletTransaction, int64, error) {
	wallet, err := r.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	var models []WalletTransactionModel
	var total int64

	if err := tx.DB(ctx, r.db).Model(&WalletTransactionModel{}).
		Where("wallet_id = ?", wallet.ID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err = tx.DB(ctx, r.db).
		Where("wallet_id = ?", wallet.ID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.WalletTransaction, 0, len(models))
	for _, m := range models {
		out = append(out, m.toDomain())
	}
	return out, total, nil
}

func (r *walletRepository) BalanceByUserID(ctx context.Context, userID string) (int64, error) {
	var model WalletModel
	err := tx.DB(ctx, r.db).Where("user_id = ?", userID).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return model.Balance, nil
}
