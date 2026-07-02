package e2e

import (
	"os"
)

// Config holds the configuration for E2E tests
type Config struct {
	// BaseURL is the base URL of the API to test (e.g., https://console.stage.redhat.com)
	BaseURL string

	// UserID is the test user's ID for x-rh-identity header
	UserID string

	// AccountID is the test user's account ID
	AccountID string

	// OrgID is the test user's organization ID
	OrgID string

	// Username is the test user's username
	Username string
}

// GetConfig returns the E2E test configuration from environment variables
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
