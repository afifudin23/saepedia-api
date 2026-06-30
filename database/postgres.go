package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/afifudin23/saepedia-api/config"
	"github.com/afifudin23/saepedia-api/pkg/logger"
)

var DB *gorm.DB

func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed open database", logger.Err(err))
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed get sql.DB", logger.Err(err))
		return err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		logger.Error("failed ping database", logger.Err(err))
		return err
	}

	logger.Info(
		"database connected",
		logger.String("host", config.AppConfig.DBHost),
		logger.String("database", config.AppConfig.DBName),
	)

	DB = db
	return nil
}
