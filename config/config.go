package config

import (
	"os"
	"strconv"
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
  config = &ChromeServiceConfig{}
  config.ServerAddr = ":8000"
  config.Test = false

  // Ignoring Clowder setup for now
  config.DbUser = os.Getenv("PGSQL_USER")
	config.DbPassword = os.Getenv("PGSQL_PASSWORD")
	config.DbHost = os.Getenv("PGSQL_HOSTNAME")
	port, _ := strconv.Atoi(os.Getenv("PGSQL_PORT"))
	config.DbPort = port
	config.DbName = os.Getenv("PGSQL_DATABASE")
	config.MetricsPort = 8080
	config.DbSSLMode = "disable"
	config.DbSSLRootCert = ""
}

// Returning chrome-service configuration
func Get() *ChromeServiceConfig {
  return config
}
