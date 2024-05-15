package util_test

import (
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/stretchr/testify/assert"
)

func TestCreateBaseLayouts(t *testing.T) {
	t.Run("Test Create Base Layouts", func(t *testing.T) {
		layouts := service.BaseTemplates
		assert.NotNil(t, layouts)
		assert.NotNil(t, layouts["landingPage"])
		assert.NotNil(t, layouts["landingPage"].TemplateConfig)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Sm)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Md)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Lg)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Xl)

		assert.Equal(t, 6, len(layouts["landingPage"].TemplateConfig.Sm.Data()))
	})
}
