package main

import (
	"fmt"

	"github.com/afifudin23/saepedia-api/pkg/helper"
	"gorm.io/gorm"
)

type seedUser struct {
	Email    string
	Password string
	IsAdmin  bool
	Roles    []string
	Balance  int64  // saldo wallet bila buyer
	Store    string // nama toko bila seller
	Products []seedProduct
}

type seedProduct struct {
	Name        string
	Description string
	Price       int64
	Stock       int
	Image       string // URL gambar (Unsplash)
}

// Users membuat akun demo (admin/seller/buyer/driver/multi-role) beserta
// wallet, toko, dan produk. Idempotent (ON CONFLICT DO NOTHING).
func Users(db *gorm.DB) error {
	users := []seedUser{
		{
			Email: "admin@seapedia.test", Password: "Admin123",
			IsAdmin: true,
		},
		// ── Sellers (punya toko + produk) ─────────────────────────
		{
			Email: "seller1@seapedia.test", Password: "Seller123",
			Roles: []string{"seller"}, Store: "Toko Sumber Rejeki",
			Products: []seedProduct{
				{"Kemeja Flanel", "Kemeja flanel lengan panjang bahan adem", 150000, 25, "https://source.unsplash.com/600x400/?flannel,shirt"},
				{"Celana Chino", "Celana chino slim fit warna khaki", 200000, 15, "https://source.unsplash.com/600x400/?chinos,pants"},
				{"Topi Baseball", "Topi baseball katun premium", 75000, 40, "https://source.unsplash.com/600x400/?cap,hat"},
				{"Jaket Denim", "Jaket denim unisex warna biru tua", 275000, 12, "https://source.unsplash.com/600x400/?denim,jacket"},
			},
		},
		{
			Email: "seller2@seapedia.test", Password: "Seller123",
			Roles: []string{"seller"}, Store: "Elektronik Jaya",
			Products: []seedProduct{
				{"Mouse Wireless", "Mouse wireless 2.4GHz hemat baterai", 95000, 50, "https://source.unsplash.com/600x400/?computer,mouse"},
				{"Keyboard Mekanik", "Keyboard mekanik RGB switch biru", 350000, 20, "https://source.unsplash.com/600x400/?mechanical,keyboard"},
				{"Headset Gaming", "Headset gaming surround 7.1", 280000, 18, "https://source.unsplash.com/600x400/?gaming,headset"},
				{"Webcam HD", "Webcam 1080p dengan mikrofon", 220000, 22, "https://source.unsplash.com/600x400/?webcam"},
			},
		},
		{
			Email: "seller3@seapedia.test", Password: "Seller123",
			Roles: []string{"seller"}, Store: "Dapur Sehat",
			Products: []seedProduct{
				{"Tumbler Stainless", "Tumbler stainless 500ml tahan panas/dingin", 85000, 60, "https://source.unsplash.com/600x400/?tumbler,bottle"},
				{"Set Pisau Dapur", "Set 5 pisau dapur anti karat", 165000, 30, "https://source.unsplash.com/600x400/?kitchen,knife"},
				{"Talenan Kayu", "Talenan kayu jati ukuran besar", 70000, 45, "https://source.unsplash.com/600x400/?cutting,board"},
			},
		},
		// ── Buyers (punya wallet) ─────────────────────────────────
		{
			Email: "buyer1@seapedia.test", Password: "Buyer123",
			Roles: []string{"buyer"}, Balance: 1000000,
		},
		{
			Email: "buyer2@seapedia.test", Password: "Buyer123",
			Roles: []string{"buyer"}, Balance: 500000,
		},
		{
			Email: "buyer3@seapedia.test", Password: "Buyer123",
			Roles: []string{"buyer"}, Balance: 2000000,
		},
		// ── Drivers ───────────────────────────────────────────────
		{
			Email: "driver1@seapedia.test", Password: "Driver123",
			Roles: []string{"driver"},
		},
		{
			Email: "driver2@seapedia.test", Password: "Driver123",
			Roles: []string{"driver"},
		},
		// ── Multi-role: buyer+seller+driver → wajib pilih role aktif ─
		{
			Email: "multi1@seapedia.test", Password: "Multi123",
			Roles: []string{"buyer", "seller", "driver"}, Balance: 500000,
			Store: "Toko Serba Ada",
			Products: []seedProduct{
				{"Botol Minum 1L", "Botol minum BPA free 1 liter", 50000, 100, "https://source.unsplash.com/600x400/?water,bottle"},
				{"Payung Lipat", "Payung lipat anti UV otomatis", 65000, 35, "https://source.unsplash.com/600x400/?umbrella"},
				{"Power Bank 10000mAh", "Power bank fast charging 10000mAh", 185000, 28, "https://source.unsplash.com/600x400/?powerbank,charger"},
			},
		},
	}

	for _, u := range users {
		if err := seedOneUser(db, u); err != nil {
			return fmt.Errorf("user %s: %w", u.Email, err)
		}
		fmt.Printf("  - user %-24s (%s)\n", u.Email, rolesLabel(u))
	}
	return nil
}

func seedOneUser(db *gorm.DB, u seedUser) error {
	hashed, err := helper.Hash(u.Password)
	if err != nil {
		return err
	}

	if err := db.Exec(
		"INSERT INTO users (email, password, is_admin) VALUES (?, ?, ?) ON CONFLICT (email) DO NOTHING",
		u.Email, hashed, u.IsAdmin,
	).Error; err != nil {
		return err
	}

	var userID string
	if err := db.Raw("SELECT id FROM users WHERE email = ?", u.Email).Scan(&userID).Error; err != nil {
		return err
	}

	for _, role := range u.Roles {
		if err := db.Exec(
			"INSERT INTO user_roles (user_id, role) VALUES (?, ?) ON CONFLICT (user_id, role) DO NOTHING",
			userID, role,
		).Error; err != nil {
			return err
		}
	}

	// Wallet untuk buyer.
	if hasRole(u.Roles, "buyer") {
		if err := db.Exec(
			"INSERT INTO wallets (user_id, balance) VALUES (?, ?) ON CONFLICT (user_id) DO NOTHING",
			userID, u.Balance,
		).Error; err != nil {
			return err
		}
	}

	// Toko + produk untuk seller.
	if u.Store != "" {
		if err := db.Exec(
			"INSERT INTO stores (user_id, name) VALUES (?, ?) ON CONFLICT (user_id) DO NOTHING",
			userID, u.Store,
		).Error; err != nil {
			return err
		}

		var storeID string
		if err := db.Raw("SELECT id FROM stores WHERE user_id = ?", userID).Scan(&storeID).Error; err != nil {
			return err
		}

		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM products WHERE store_id = ?", storeID).Scan(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			for _, p := range u.Products {
				if err := db.Exec(
					"INSERT INTO products (store_id, name, description, price, stock, image_url) VALUES (?, ?, ?, ?, ?, ?)",
					storeID, p.Name, p.Description, p.Price, p.Stock, p.Image,
				).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func rolesLabel(u seedUser) string {
	if u.IsAdmin {
		return "admin"
	}
	out := ""
	for i, r := range u.Roles {
		if i > 0 {
			out += ","
		}
		out += r
	}
	return out
}
