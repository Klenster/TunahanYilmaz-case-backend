package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config uygulama genelinde kullanılan ayarları tutar.
// Tüm değerler .env dosyasından veya ortam değişkenlerinden okunur.
type Config struct {
	AppPort        string
	AppEnv         string
	SecretKey      string
	JWTExpiryHours int
	DBPath         string
	AdminEmail     string
	AdminPassword  string
}

func Load() *Config {
	for _, p := range []string{".env", "../../.env", "../../../.env"} {
		if err := godotenv.Load(p); err == nil {
			break
		}
	}

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))

	return &Config{
		AppPort:        getEnv("APP_PORT", "8000"),
		AppEnv:         getEnv("APP_ENV", "development"),
		SecretKey:      getEnv("APP_SECRET_KEY", ""),
		JWTExpiryHours: expiryHours,
		DBPath:         getEnv("DB_PATH", "./dpp.db"),
		AdminEmail:     getEnv("ADMIN_EMAIL", "admin@passportx.com"),
		AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}