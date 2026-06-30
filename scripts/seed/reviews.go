package main

import (
	"fmt"

	"gorm.io/gorm"
)

// Reviews mengisi 10 review aplikasi (app_reviews). Idempotent: dilewati bila
// tabel sudah berisi data (app_reviews tidak punya kolom unik).
func Reviews(db *gorm.DB) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM app_reviews").Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("  - app_reviews sudah ada, dilewati")
		return nil
	}

	reviews := []struct {
		Name    string
		Rating  int
		Comment string
	}{
		{"Andi Saputra", 5, "Aplikasinya mulus banget, checkout cepat!"},
		{"Bunga Lestari", 4, "Tampilan rapi, tinggal tambah metode bayar lain."},
		{"Citra Dewi", 5, "Suka banget bisa pilih banyak peran dalam satu akun."},
		{"Dimas Pratama", 4, "Pengiriman gampang dilacak, mantap."},
		{"Eka Putri", 5, "Voucher dan promonya membantu hemat."},
		{"Fajar Nugroho", 3, "Cukup baik, tapi loading katalog bisa lebih cepat."},
		{"Gita Anggraini", 5, "Top up dompet praktis, langsung kepakai."},
		{"Hadi Wijaya", 4, "Driver dapat job dengan jelas, alurnya enak."},
		{"Indah Permata", 5, "Marketplace lokal yang menjanjikan!"},
		{"Joko Susilo", 4, "Fitur review tanpa harus login itu plus."},
	}

	for _, r := range reviews {
		if err := db.Exec(
			"INSERT INTO app_reviews (reviewer_name, rating, comment) VALUES (?, ?, ?)",
			r.Name, r.Rating, r.Comment,
		).Error; err != nil {
			return err
		}
		fmt.Printf("  - review oleh %s (%d★)\n", r.Name, r.Rating)
	}
	return nil
}
