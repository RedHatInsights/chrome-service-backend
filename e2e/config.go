package e2e

import (
	"os"
)

const APIBasePath = "/api/chrome-service/v1"

type Config struct {
	BaseURL   string
	UserID    string
	AccountID string
	OrgID     string
	Username  string
}

func GetConfig() *Config {
	return &Config{
		BaseURL:   getEnvOrDefault("E2E_BASE_URL", "http://localhost:8000"),
		UserID:    getEnvOrDefault("E2E_USER_ID", "test-user-123"),
		AccountID: getEnvOrDefault("E2E_ACCOUNT_ID", "123456"),
		OrgID:     getEnvOrDefault("E2E_ORG_ID", "654321"),
		Username:  getEnvOrDefault("E2E_USERNAME", "testuser"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
