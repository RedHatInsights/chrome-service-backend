package e2e

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// Load .env file if it exists (for local testing)
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using environment variables")
	}

	// Run all tests
	exitCode := m.Run()

	os.Exit(exitCode)
}
