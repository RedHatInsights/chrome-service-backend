package util_test

import (
	"encoding/json"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/stretchr/testify/assert"
)

func TestCreateBaseLayouts(t *testing.T) {
	t.Run("Test Create Base Layouts", func(t *testing.T) {
		layouts := util.BaseTemplates
		assert.NotNil(t, layouts)
		assert.NotNil(t, layouts["landingPage"])
		assert.NotNil(t, layouts["landingPage"].TemplateConfig)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Sx)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Md)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Lg)
		assert.NotNil(t, layouts["landingPage"].TemplateConfig.Xl)

		var gridItems []map[string]interface{}
		json.Unmarshal(layouts["landingPage"].TemplateConfig.Sx, &gridItems)

		assert.Equal(t, 6, len(gridItems))
	})
}
