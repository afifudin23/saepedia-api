package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppPort    string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	AccessKey  string
}

var AppConfig Config

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func requireEnv(key string, missing *[]string) string {
	value := os.Getenv(key)
	if value == "" {
		*missing = append(*missing, key)
	}
	return value
}

func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env tidak ditemukan, pakai environment variable sistem")
	}

	var missing []string

	AppConfig = Config{
		AppName: getEnv("APP_NAME", "SEAPEDIA API"),
		AppPort: getEnv("APP_PORT", "5000"),
		DBHost:  getEnv("DB_HOST", "localhost"),
		DBPort:  getEnv("DB_PORT", "5432"),
		DBUser:  getEnv("DB_USER", "postgres"),
		DBName:  getEnv("DB_NAME", "seapedia_db"),

		DBPassword: requireEnv("DB_PASSWORD", &missing),
		AccessKey:  requireEnv("ACCESS_KEY", &missing),
	}

	if len(missing) > 0 {
		log.Fatalf("config: required env not set: %s", strings.Join(missing, ", "))
	}
}
