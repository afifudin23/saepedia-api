package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrRoleAlreadyOwned = errors.New("role already owned by user")
	ErrInvalidRole      = errors.New("invalid role")
)

// User adalah entity bisnis. Identitas akun = email (unik). Roles berisi seluruh
// role NON-ADMIN yang dimiliki user. Role aktif (untuk authorization) disimpan
// terpisah di JWT per sesi.
type User struct {
	ID        string
	Email     string
	Password  string
	IsAdmin   bool
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasRole mengecek apakah user memiliki role tertentu.
// Admin dianggap memiliki role "admin".
func (u *User) HasRole(role string) bool {
	if role == "admin" {
		return u.IsAdmin
	}
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type UserRepository interface {
	// Create menyimpan user beserta role-nya dalam satu transaksi.
	Create(ctx context.Context, user *User, roles []string) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	AddRole(ctx context.Context, userID, role string) error
	// CountAll & ListAll dipakai untuk Admin monitoring di level berikutnya.
	CountAll(ctx context.Context) (int64, error)
}
