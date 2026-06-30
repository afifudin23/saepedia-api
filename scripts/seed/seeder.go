package main

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Seeder adalah satu unit seeding (mis. user, discount).
type Seeder struct {
	Name string
	Run  func(db *gorm.DB) error
}

// registry mengembalikan semua seeder dalam URUTAN dependency (user dulu, dst).
func registry() []Seeder {
	// Urutan penting: user (toko, produk, wallet) & address dulu sebelum order.
	return []Seeder{
		{Name: "user", Run: Users},
		{Name: "discount", Run: Discounts},
		{Name: "review", Run: Reviews},
		{Name: "address", Run: Addresses},
		{Name: "order", Run: Orders},
	}
}

func seederNames() []string {
	all := registry()
	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, s.Name)
	}
	return names
}

// getSeeder mencari seeder berdasarkan nama (lenient: huruf besar/kecil & bentuk
// jamak diterima, mis. "user", "Users", "users").
func getSeeder(name string) (Seeder, bool) {
	norm := normalize(name)
	for _, s := range registry() {
		if s.Name == norm {
			return s, true
		}
	}
	return Seeder{}, false
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.TrimSuffix(s, "s")
}

// runAll menjalankan seluruh seeder secara berurutan.
func runAll(db *gorm.DB) error {
	for _, s := range registry() {
		fmt.Printf("Menjalankan seeder %q...\n", s.Name)
		if err := s.Run(db); err != nil {
			return fmt.Errorf("seeder %q: %w", s.Name, err)
		}
	}
	return nil
}
