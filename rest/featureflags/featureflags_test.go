package featureflags

import (
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/stretchr/testify/assert"
)

func TestBasicFeatureFlagConnection(t *testing.T) {
	t.Run("Test accessible unleash server", func(t *testing.T) {
		Init(util.SetupTestConfig())
		assert.NotNil(t, GetClient())
	})
	t.Run("Test disabled flag is disabled", func(t *testing.T) {
		assert.False(t, IsEnabled("unit-test.false"))
	})
	t.Run("Test enabled flag is enabled", func(t *testing.T) {
		assert.True(t, IsEnabled("unit-test.true"))
		Close()
	})
}

func TestBrokenFeatureFlagConnection(t *testing.T) {
	cfg := util.SetupTestConfig()
	cfg.FeatureFlagConfig.FullURL = "gohawaii.com"
	t.Run("Connect to vacation URL", func(t *testing.T) {
		Init(cfg)
		assert.Empty(t, GetClient())
		Close()
	})
}

func TestEmptyFeatureFlagConfig(t *testing.T) {
	cfg := util.SetupTestConfig()
	cfg.FeatureFlagConfig = config.FeatureFlagsConfig{}
	t.Run("Test missing FeatureFlag config", func(t *testing.T) {
		Init(cfg)
		assert.Nil(t, GetClient())
		// True flags should be false if the server cannot be reached
		assert.False(t, IsEnabled("unit-test.true"))
		Close()
	})
}

func TestPersistentFeatureFlagConnection(t *testing.T) {
	t.Run("Setup basic connection", func(t *testing.T) {
		Init(util.SetupTestConfig())
		assert.NotNil(t, GetClient())
	})
	t.Run("Ensure client object is persistent", func(t *testing.T) {
		assert.NotNil(t, GetClient())
		assert.False(t, IsEnabled("fake.flag"))
		Close()
	})
	t.Run("Client is nil after close", func(t *testing.T) {
		assert.Nil(t, GetClient())
	})
}
