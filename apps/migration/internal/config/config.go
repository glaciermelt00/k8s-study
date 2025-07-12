package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	dbHost := getEnv("DB_HOST", "postgres-headless")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	dbName := getEnv("DB_NAME", "postgres")
	
	// SSLモードを環境変数から取得（デフォルトは require）
	// Kubernetes内部通信の場合は DB_SSLMODE=disable を設定可能
	sslMode := getEnv("DB_SSLMODE", "require")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

	return &Config{
		DatabaseURL: databaseURL,
		Port:        getEnv("PORT", "8080"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
