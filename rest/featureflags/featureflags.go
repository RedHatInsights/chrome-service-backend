package featureflags

import (
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/Unleash/unleash-client-go/v3"
	"github.com/sirupsen/logrus"
)

// This is a wrapper struct to setup our own requests to avoid calling
// a nil client. It's all a bit deranged, but I cannot get the library to
// guarantee any sane defaults
type FFClient struct {
	unleashClient *unleash.Client
}

var ffClient *FFClient

func newClientWrapper(cfg *config.ChromeServiceConfig) (*unleash.Client, error) {
	client, err := unleash.NewClient(
		unleash.WithListener(&unleash.DebugListener{}),
		unleash.WithAppName("chrome-service"),
		unleash.WithUrl(cfg.FeatureFlagConfig.FullURL),
		unleash.WithCustomHeaders(http.Header{"Authorization": {cfg.FeatureFlagConfig.ClientAccessToken}}),
	)
	if err != nil {
		return nil, err
	}

	// Reading this channel makes me a bit nervous and I'd like to do more
	// testing to see if we need to empty all the channels
	select {
	case errs := <-client.Errors():
		client.Close()
		return nil, errs
	default:
		return client, nil
	}

}

func New(cfg *config.ChromeServiceConfig) error {
	// If there is already an established connection, don't make a new one
	if ffClient != nil {
		return nil
	}
	c, err := newClientWrapper(cfg)
	if err != nil {
		ffClient = nil
		logrus.Infof("Unable to contact unleash server due to: %v", err)
		return err
	}
	ffClient = &FFClient{
		unleashClient: c,
	}
	return nil
}

// Wrap the unleash to avoid having to do
// if (unleashClient != nil && unleashClient.IsEnabled("FeatureFlag"))
// every single call
func IsEnabled(flag string) bool {
	if ffClient != nil {
		return ffClient.unleashClient.IsEnabled(flag)
	} else {
		return false
	}
}

func Close() {
	if ffClient != nil {
		ffClient.unleashClient.Close()
		ffClient = nil
	}
}

// Called before main() and when the library is imported
func Init(cfg *config.ChromeServiceConfig) {
	err := New(cfg)
	if err != nil {
		logrus.Infoln("all feature flags are set to false")
	}
}

func GetClient() *FFClient {
	return ffClient
}
