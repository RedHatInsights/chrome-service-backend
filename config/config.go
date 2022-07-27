package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ChromeServiceConfig struct {
	ServerAddr      string
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

var config *ChromeServiceConfig

func Init() {
	godotenv.Load()
	options := &ChromeServiceConfig{}
	options.ServerAddr = ":8000"
	options.Test = false

	// Ignoring Clowder setup for now
	options.DbUser = os.Getenv("PGSQL_USER")
	options.DbPassword = os.Getenv("PGSQL_PASSWORD")
	options.DbHost = os.Getenv("PGSQL_HOSTNAME")
	port, _ := strconv.Atoi(os.Getenv("PGSQL_PORT"))
	options.DbPort = port
	options.DbName = os.Getenv("PGSQL_DATABASE")
	options.MetricsPort = 8080
	options.DbSSLMode = "disable"
	options.DbSSLRootCert = ""

	config = options
}

// Returning chrome-service configuration
func Get() *ChromeServiceConfig {
	return config
}
