package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

type ChromeServiceConfig struct {
	WebPort         int
	OpenApiSpecPath string
	DbHost          string
	DbUser          string
	DbPassword      string
	DbPort          int
	DbName          string
	MetricsPort     int
	Test            bool
	DbSSLMode       string
	DbSSLRootCert   string
}

const RdsCaLocation = "/app/rdsca.cert"

func (c *ChromeServiceConfig) getCert(cfg *clowder.AppConfig) string {
	cert := ""
	if cfg.Database.SslMode != "verify-full" {
		return cert
	}
	if cfg.Database.RdsCa != nil {
		err := os.WriteFile(RdsCaLocation, []byte(*cfg.Database.RdsCa), 0644)
		if err != nil {
			panic(err)
		}
		cert = RdsCaLocation
	}
	return cert
}

var config *ChromeServiceConfig

func Init() {
	godotenv.Load()
	options := &ChromeServiceConfig{}

	if clowder.IsClowderEnabled() {
		cfg := clowder.LoadedConfig
		options.DbName = cfg.Database.Name
		options.DbHost = cfg.Database.Hostname
		options.DbPort = cfg.Database.Port
		options.DbUser = cfg.Database.Username
		options.DbPassword = cfg.Database.Password
		options.MetricsPort = cfg.MetricsPort
		options.DbSSLMode = cfg.Database.SslMode
		options.DbSSLRootCert = options.getCert(cfg)
		options.WebPort = *cfg.PublicPort
	} else {
		options.WebPort = 8000
		options.Test = false

		// Ignoring Clowder setup for now
		options.DbUser = os.Getenv("PGSQL_USER")
		options.DbPassword = os.Getenv("PGSQL_PASSWORD")
		options.DbHost = os.Getenv("PGSQL_HOSTNAME")
		port, _ := strconv.Atoi(os.Getenv("PGSQL_PORT"))
		options.DbPort = port
		options.DbName = os.Getenv("PGSQL_DATABASE")
		options.MetricsPort = 9000
		options.DbSSLMode = "disable"
		options.DbSSLRootCert = ""
	}

	config = options
}

// Returning chrome-service configuration
func Get() *ChromeServiceConfig {
	return config
}
