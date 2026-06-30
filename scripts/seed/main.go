// Seeder data demo SEAPEDIA.
//
//	go run ./scripts/seed              # jalankan SEMUA seeder
//	go run ./scripts/seed data=user    # jalankan seeder user saja
//	go run ./scripts/seed data=discount
//
// Seeder bersifat idempotent (aman dijalankan berulang).
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/afifudin23/saepedia-api/config"
	"github.com/afifudin23/saepedia-api/database"
)

func main() {
	// Ambil argumen data=<nama> bila ada.
	only := ""
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "data=") {
			only = strings.TrimPrefix(arg, "data=")
		}
	}

	// Validasi nama seeder sebelum menyentuh database.
	if only != "" {
		if _, ok := getSeeder(only); !ok {
			log.Fatalf("seeder %q tidak ditemukan. Tersedia: %s", only, strings.Join(seederNames(), ", "))
		}
	}

	config.LoadConfig()
	if err := database.Connect(); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if only == "" {
		fmt.Println("Menjalankan SEMUA seeder...")
		if err := runAll(database.DB); err != nil {
			log.Fatalf("seed: %v", err)
		}
	} else {
		s, _ := getSeeder(only)
		fmt.Printf("Menjalankan seeder %q...\n", s.Name)
		if err := s.Run(database.DB); err != nil {
			log.Fatalf("seeder %q: %v", s.Name, err)
		}
	}

	fmt.Println("\nSeed selesai. Login memakai email (akun demo & password ada di README).")
}
