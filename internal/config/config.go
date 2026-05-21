package config

import (
	"os"
	"time"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8083"), // Порт по умолчанию для заказов
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mixfood_orders?sslmode=disable"),
		JWTSecret:   jwtSecret,
		AccessTTL:   15 * time.Minute,
		RefreshTTL:  30 * 24 * time.Hour,
	}
}
