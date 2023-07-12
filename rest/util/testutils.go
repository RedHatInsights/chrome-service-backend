package util

import (
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/config"
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
