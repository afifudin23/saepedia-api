// Package tx menyediakan transaction manager berbasis context.
//
// Tujuannya: business logic tetap di USECASE (yang memutuskan urutan & aturan),
// sedangkan REPOSITORY hanya menyediakan operasi data. Usecase membungkus
// beberapa pemanggilan repo dalam satu transaksi DB lewat Manager.Do, dan setiap
// repo otomatis ikut transaksi yang sama karena membaca *gorm.DB dari context
// via DB().
package tx

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

// Manager menjalankan beberapa operasi repo dalam satu transaksi DB.
type Manager struct {
	db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// Do menjalankan fn di dalam satu transaksi. Bila fn mengembalikan error,
// seluruh perubahan di-rollback. Repo di dalam fn memakai transaksi yang sama.
func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(txDB *gorm.DB) error {
		return fn(context.WithValue(ctx, ctxKey{}, txDB))
	})
}

// DB mengembalikan handle DB yang aktif untuk context ini: transaksi bila sedang
// di dalam Manager.Do, atau koneksi biasa (fallback) bila tidak.
func DB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if txDB, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return txDB
	}
	return fallback.WithContext(ctx)
}
