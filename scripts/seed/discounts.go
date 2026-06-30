package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type seedDiscount struct {
	Code        string
	Kind        string // voucher / promo
	Type        string // percent / fixed
	Value       int64
	MaxDiscount int64 // cap untuk percent (0 = tanpa cap)
	MinSpend    int64
	UsageLimit  *int // hanya voucher (promo nil)
}

// Discounts membuat 10 kode diskon demo (5 voucher + 5 promo). Idempotent.
func Discounts(db *gorm.DB) error {
	expires := time.Now().AddDate(1, 0, 0) // berlaku 1 tahun

	limit := func(n int) *int { return &n }

	discounts := []seedDiscount{
		// ── Voucher (punya kuota pemakaian) ───────────────────────
		{"SEAPEDIA10", "voucher", "percent", 10, 25000, 50000, limit(100)},
		{"GROCERY15", "voucher", "percent", 15, 50000, 100000, limit(75)},
		{"NEWUSER20", "voucher", "percent", 20, 30000, 50000, limit(200)},
		{"FLASH25", "voucher", "percent", 25, 100000, 150000, limit(50)},
		{"HEMAT50K", "voucher", "fixed", 50000, 0, 300000, limit(40)},
		// ── Promo (tanpa kuota) ───────────────────────────────────
		{"HEMAT5K", "promo", "fixed", 5000, 0, 30000, nil},
		{"POTONG10K", "promo", "fixed", 10000, 0, 75000, nil},
		{"POTONG20K", "promo", "fixed", 20000, 0, 150000, nil},
		{"PROMO10", "promo", "percent", 10, 40000, 60000, nil},
		{"GAJIAN", "promo", "percent", 15, 60000, 100000, nil},
	}

	for _, d := range discounts {
		err := db.Exec(`
			INSERT INTO discounts (code, kind, discount_type, discount_value, max_discount, min_spend, expires_at, usage_limit)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (code) DO NOTHING`,
			d.Code, d.Kind, d.Type, d.Value, d.MaxDiscount, d.MinSpend, expires, d.UsageLimit,
		).Error
		if err != nil {
			return fmt.Errorf("discount %s: %w", d.Code, err)
		}
		fmt.Printf("  - %-7s %s\n", d.Kind, d.Code)
	}
	return nil
}
