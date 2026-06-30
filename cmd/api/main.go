// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/afifudin23/saepedia-api/config"
	"github.com/afifudin23/saepedia-api/database"
	swaggerdocs "github.com/afifudin23/saepedia-api/docs/swagger" // generated swagger docs
	"github.com/afifudin23/saepedia-api/internal/router"
	"github.com/afifudin23/saepedia-api/pkg/logger"
)

// @title                       SEAPEDIA API
// @version                     1.0
// @description                 Backend API marketplace multi-role (Admin, Seller, Buyer, Driver) — COMPFEST 18.
// @description                 Versi yang tampil di UI diambil dari config.Version (di-set saat runtime).
// @description                 Login pakai email. Otorisasi mengikuti ACTIVE ROLE pada JWT (pilih via /auth/select-role).
// @termsOfService              http://swagger.io/terms/
// @contact.name                SEAPEDIA Dev
// @BasePath                    /api/v1
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Masukkan token dengan format: Bearer <token>
func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}

func run() error {
	config.LoadConfig()

	// Versi Swagger diambil dari config (satu sumber kebenaran).
	swaggerdocs.SwaggerInfo.Version = config.Version

	logger.Init()
	defer logger.Sync()

	if err := database.Connect(); err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	r := router.Setup(database.DB, config.AppConfig.AccessKey)

	logger.Info("server started", logger.String("port", config.AppConfig.AppPort))
	if err := r.Run(":" + config.AppConfig.AppPort); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}
