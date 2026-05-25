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

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(key + " is required")
	}
	return value
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8083"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mixfood_orders?sslmode=disable"),
		JWTSecret:   getEnvRequired("JWT_SECRET"),
		AccessTTL:   15 * time.Minute,
		RefreshTTL:  30 * 24 * time.Hour,
	}
}
