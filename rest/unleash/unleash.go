package unleash

import (
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/Unleash/unleash-client-go/v3"
	"github.com/sirupsen/logrus"
)

func newClientWrapper(cfg *config.ChromeServiceConfig) (*unleash.Client, error) {
	client, err := unleash.NewClient(
		unleash.WithListener(&unleash.DebugListener{}),
		unleash.WithAppName("chrome-service"),
		unleash.WithUrl(cfg.FeatureFlagConfig.FullURL),
		unleash.WithCustomHeaders(http.Header{"Authorization": {cfg.FeatureFlagConfig.ClientAccessToken}}),
	)
	if err != nil {
		client.Close()
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

// This is a wrapper struct to setup our own requests to avoid calling
// a nil client. It's all a bit deranged, but I cannot get the library to
// guarantee any sane defaults
type FFClient struct {
	unleashClient *unleash.Client
}

func New(cfg *config.ChromeServiceConfig) (*FFClient, error) {
	c, err := newClientWrapper(cfg)
	if err != nil {
		logrus.Infof("Unable to contact unleash server due to: %v", err)
		return nil, err
	}
	ffc := &FFClient{
		unleashClient: c,
	}
	return ffc, nil
}

// Wrap the unleash to avoid having to do
// if (unleashClient != nil && unleashClient.IsEnabled("FeatureFlag"))
// every single call
func (ffClient *FFClient) IsEnabled(flag string) bool {
	if ffClient != nil {
		return ffClient.unleashClient.IsEnabled(flag)
	} else {
		return false
	}
}

func (ffClient *FFClient) Close() {
	ffClient.unleashClient.Close()
}
