package domain

import (
	"context"
	"errors"
	"time"

	userdomain "github.com/afifudin23/saepedia-api/internal/user/domain"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrRoleNotOwned       = errors.New("active role not owned by user")
)

type RegisterInput struct {
	Email    string
	Password string
	Roles    []string // role non-admin yang didaftarkan
}

type LoginInput struct {
	Email    string
	Password string
}

// AuthResult dikembalikan setelah register/login/select-role.
// NeedRoleSelection true berarti user punya >1 role non-admin dan HARUS memilih
// role aktif dulu (token tetap diberikan, tapi belum bisa akses dashboard privat).
type AuthResult struct {
	User              *userdomain.User
	Token             string
	ActiveRole        string
	NeedRoleSelection bool
}

// BalanceSummary adalah ringkasan finansial lintas role untuk satu akun (email).
// Wallet balance nyata sudah aktif (Level 3); seller income & driver earnings
// adalah placeholder yang akan diisi di level berikutnya.
type BalanceSummary struct {
	WalletBalance  int64 `json:"wallet_balance"`
	SellerIncome   int64 `json:"seller_income"`
	DriverEarnings int64 `json:"driver_earnings"`
}

// WalletReader memberi akses baca saldo dompet tanpa auth tahu detail wallet.
type WalletReader interface {
	BalanceByUserID(ctx context.Context, userID string) (int64, error)
}

// TokenRevoker mengelola denylist token untuk logout (invalidasi JWT).
type TokenRevoker interface {
	Revoke(ctx context.Context, jti string, exp time.Time) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, in RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, in LoginInput) (*AuthResult, error)
	SelectRole(ctx context.Context, userID, role string) (*AuthResult, error)
	Logout(ctx context.Context, jti string, exp time.Time) error
	Me(ctx context.Context, userID string) (*userdomain.User, error)
	BalanceSummary(ctx context.Context, userID string) (*BalanceSummary, error)
}
