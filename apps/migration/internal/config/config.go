package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() (*Config, error) {
	// 環境変数の取得
	dbHost := getEnv("DB_HOST", "postgres-headless")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := getEnv("DB_NAME", "postgres")
	sslMode := getEnv("DB_SSLMODE", "require")

	// 入力検証
	// ホスト名の検証
	if dbHost == "" {
		return nil, fmt.Errorf("DB_HOST cannot be empty")
	}

	// ポート番号の検証
	port, err := strconv.Atoi(dbPort)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid DB_PORT: %s", dbPort)
	}

	// ユーザー名の検証
	if dbUser == "" {
		return nil, fmt.Errorf("DB_USER cannot be empty")
	}

	// パスワードの検証
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	// データベース名の検証
	if dbName == "" {
		return nil, fmt.Errorf("DB_NAME cannot be empty")
	}

	// SSLモードの検証
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[sslMode] {
		return nil, fmt.Errorf("invalid DB_SSLMODE: %s", sslMode)
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, sslMode)

	// APIポートの検証
	apiPort := getEnv("PORT", "8080")
	apiPortNum, err := strconv.Atoi(apiPort)
	if err != nil || apiPortNum < 1 || apiPortNum > 65535 {
		return nil, fmt.Errorf("invalid PORT: %s", apiPort)
	}

	return &Config{
		DatabaseURL: databaseURL,
		Port:        apiPort,
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
