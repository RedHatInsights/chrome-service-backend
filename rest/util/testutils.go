package util

import (
	"fmt"
	"os"
	"regexp"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/joho/godotenv"
)

const ProjectName = "chrome-service-backend"

func SetupTestConfig() *config.ChromeServiceConfig {
	cfg := &config.ChromeServiceConfig{}
	cfg.DbHost = "localhost"
	cfg.DbUser = "chrome"
	cfg.DbPassword = "chrome"
	cfg.DbPort = 5432
	cfg.DbName = "chrome"
	cfg.FeatureFlagConfig.ClientAccessToken = "default:development.unleash-insecure-api-token"
	cfg.FeatureFlagConfig.Hostname = "localhost"
	cfg.FeatureFlagConfig.Scheme = "http"
	cfg.FeatureFlagConfig.Port = 4242
	cfg.FeatureFlagConfig.FullURL = fmt.Sprintf("%s://%s:%d/api/", cfg.FeatureFlagConfig.Scheme, cfg.FeatureFlagConfig.Hostname, cfg.FeatureFlagConfig.Port)
	return cfg
}

// TODO: Break test config out into env files in the future
// LoadEnv loads env vars from .env
func LoadEnv() {
	re := regexp.MustCompile(`^(.*` + ProjectName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/.env.test`)
	if err != nil {
		fmt.Println(err)
	}
}
