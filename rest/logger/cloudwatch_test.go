package logger

import (
	"os"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCloudWatchFallbackAndFlush(t *testing.T) {
	standardLogger := logrus.StandardLogger()
	oldOutput, oldFormatter := standardLogger.Out, standardLogger.Formatter
	t.Cleanup(func() {
		standardLogger.SetOutput(oldOutput)
		standardLogger.SetFormatter(oldFormatter)
	})

	SetupCloudWatch(&config.ChromeServiceConfig{})
	assert.Same(t, os.Stdout, standardLogger.Out)
	_, isJSON := standardLogger.Formatter.(*logrus.JSONFormatter)
	assert.True(t, isJSON)
	FlushCloudWatch()
}
