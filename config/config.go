package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

type KafkaSSLCfg struct {
	KafkaCA       string
	KafkaUsername string
	KafkaPassword string
	SASLMechanism string
	Protocol      string
}

type KafkaCfg struct {
	KafkaBrokers   []string
	KafkaTopics    []string
	KafkaSSlConfig KafkaSSLCfg
}

type IntercomConfig struct {
	fallback      string
	openshift     string
	openshift_dev string
	hacCore       string
}

type FeatureFlagsConfig struct {
	ClientAccessToken string
	Hostname          string
	Port              int
	Scheme            string
	FullURL           string
	// ONLY for local use, Clowder won't populate this
	AdminToken string
}

type ChromeServiceConfig struct {
	WebPort           int
	OpenApiSpecPath   string
	DbHost            string
	DbUser            string
	DbPassword        string
	DbPort            int
	DbName            string
	MetricsPort       int
	Test              bool
	LogLevel          string
	DbSSLMode         string
	DbSSLRootCert     string
	KafkaConfig       KafkaCfg
	IntercomConfig    IntercomConfig
	FeatureFlagConfig FeatureFlagsConfig
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

func init() {
	godotenv.Load()
	options := &ChromeServiceConfig{}

	// Log level will default to "info". Level should be one of
	// info or debug
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "info"
	}
	options.LogLevel = level

	if clowder.IsClowderEnabled() {
		cfg := clowder.LoadedConfig
		options.DbName = cfg.Database.Name
		options.DbHost = cfg.Database.Hostname
		options.DbPort = cfg.Database.Port
		options.DbUser = cfg.Database.Username
		options.DbPassword = cfg.Database.Password
		options.MetricsPort = cfg.MetricsPort
		options.WebPort = *cfg.PublicPort
		options.DbSSLMode = cfg.Database.SslMode
		options.DbSSLRootCert = options.getCert(cfg)

		if cfg.FeatureFlags != nil {
			options.FeatureFlagConfig.ClientAccessToken = *cfg.FeatureFlags.ClientAccessToken
			options.FeatureFlagConfig.Hostname = cfg.FeatureFlags.Hostname
			options.FeatureFlagConfig.Scheme = string(cfg.FeatureFlags.Scheme)
			options.FeatureFlagConfig.Port = cfg.FeatureFlags.Port
			options.FeatureFlagConfig.FullURL = fmt.Sprintf("%s://%s:%d/api/", options.FeatureFlagConfig.Scheme, options.FeatureFlagConfig.Hostname, options.FeatureFlagConfig.Port)
		}

		broker := cfg.Kafka.Brokers[0]
		// pass all required topics names
		for _, topic := range cfg.Kafka.Topics {
			options.KafkaConfig.KafkaTopics = append(options.KafkaConfig.KafkaTopics, topic.Name)
		}

		options.KafkaConfig.KafkaBrokers = clowder.KafkaServers
		// Kafka SSL Config
		if broker.Authtype != nil {
			options.KafkaConfig.KafkaSSlConfig.KafkaUsername = *broker.Sasl.Username
			options.KafkaConfig.KafkaSSlConfig.KafkaPassword = *broker.Sasl.Password
			options.KafkaConfig.KafkaSSlConfig.SASLMechanism = *broker.Sasl.SaslMechanism
			options.KafkaConfig.KafkaSSlConfig.Protocol = *broker.Sasl.SecurityProtocol
		}

		if broker.Cacert != nil {
			caPath, err := cfg.KafkaCa(broker)
			if err != nil {
				panic(fmt.Sprintln("Kafka CA failed to write", err))
			}
			options.KafkaConfig.KafkaSSlConfig.KafkaCA = caPath
		}
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
		options.KafkaConfig = KafkaCfg{
			KafkaTopics:  []string{"platform.chrome"},
			KafkaBrokers: []string{"localhost:9092"},
		}

		options.FeatureFlagConfig.ClientAccessToken = os.Getenv("UNLEASH_API_TOKEN")
		// Only for local use to seed the database, does not work in Clowder.
		options.FeatureFlagConfig.AdminToken = os.Getenv("UNLEASH_ADMIN_TOKEN")
		options.FeatureFlagConfig.Hostname = "0.0.0.0"
		options.FeatureFlagConfig.Scheme = "http"
		options.FeatureFlagConfig.Port = 4242
		options.FeatureFlagConfig.FullURL = fmt.Sprintf("%s://%s:%d/api/", options.FeatureFlagConfig.Scheme, options.FeatureFlagConfig.Hostname, options.FeatureFlagConfig.Port)

	}

	// env variables from .env or pod env variables
	options.IntercomConfig = IntercomConfig{
		fallback:      os.Getenv("INTERCOM_DEFAULT"),
		openshift:     os.Getenv("INTERCOM_OPENSHIFT"),
		openshift_dev: os.Getenv("INTERCOM_OPENSHIFT_DEV"),
		hacCore:       os.Getenv("INTERCOM_HAC_CORE"),
	}
	config = options
}

// Returning chrome-service configuration
func Get() *ChromeServiceConfig {
	return config
}
