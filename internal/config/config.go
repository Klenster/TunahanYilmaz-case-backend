package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config uygulama genelinde kullanılan ayarları tutar.
// Tüm değerler .env dosyasından veya ortam değişkenlerinden okunur.
type Config struct {
	AppPort       string
	AppEnv        string
	SecretKey     string
	JWTExpiryHours int
	DBPath        string
	AdminEmail    string
	AdminPassword string
}

func Load() *Config {
	// .env dosyası yoksa hata vermez, ortam değişkenlerini kullanır
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))

	return &Config{
		AppPort:        getEnv("APP_PORT", "8000"),
		AppEnv:         getEnv("APP_ENV", "development"),
		SecretKey:      getEnv("APP_SECRET_KEY", "change-me-in-production"),
		JWTExpiryHours: expiryHours,
		DBPath:         getEnv("DB_PATH", "./dpp.db"),
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@dpp.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
