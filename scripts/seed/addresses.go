package main

import (
	"fmt"

	"gorm.io/gorm"
)

// Addresses mengisi 10 alamat pengiriman untuk para buyer. Idempotent: dilewati
// bila tabel sudah berisi data.
func Addresses(db *gorm.DB) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM addresses").Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("  - addresses sudah ada, dilewati")
		return nil
	}

	addresses := []struct {
		Email     string
		Recipient string
		Phone     string
		Full      string
		Primary   bool
	}{
		{"buyer1@seapedia.test", "Andi Saputra", "081200000001", "Jl. Merdeka No. 1, Bandung", true},
		{"buyer1@seapedia.test", "Andi (Kantor)", "081200000002", "Jl. Asia Afrika No. 10, Bandung", false},
		{"buyer1@seapedia.test", "Andi (Rumah Ortu)", "081200000003", "Jl. Cihampelas No. 5, Bandung", false},
		{"buyer2@seapedia.test", "Bunga Lestari", "081200000004", "Jl. Sudirman No. 21, Jakarta", true},
		{"buyer2@seapedia.test", "Bunga (Kost)", "081200000005", "Jl. Kebon Jeruk No. 8, Jakarta", false},
		{"buyer3@seapedia.test", "Citra Dewi", "081200000006", "Jl. Diponegoro No. 12, Surabaya", true},
		{"buyer3@seapedia.test", "Citra (Kantor)", "081200000007", "Jl. Tunjungan No. 3, Surabaya", false},
		{"buyer3@seapedia.test", "Citra (Gudang)", "081200000008", "Jl. Rungkut Industri No. 9, Surabaya", false},
		{"multi1@seapedia.test", "Dimas Pratama", "081200000009", "Jl. Malioboro No. 7, Yogyakarta", true},
		{"multi1@seapedia.test", "Dimas (Toko)", "081200000010", "Jl. Kaliurang No. 15, Yogyakarta", false},
	}

	for _, a := range addresses {
		var userID string
		if err := db.Raw("SELECT id FROM users WHERE email = ?", a.Email).Scan(&userID).Error; err != nil {
			return err
		}
		if userID == "" {
			continue // user belum di-seed
		}
		if err := db.Exec(
			"INSERT INTO addresses (user_id, recipient_name, phone, full_address, is_primary) VALUES (?, ?, ?, ?, ?)",
			userID, a.Recipient, a.Phone, a.Full, a.Primary,
		).Error; err != nil {
			return err
		}
		fmt.Printf("  - alamat %s → %s\n", a.Email, a.Recipient)
	}
	return nil
}
