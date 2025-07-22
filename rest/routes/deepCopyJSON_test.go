package routes

import (
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/stretchr/testify/assert"
)

func TestDeepCopyJSON(t *testing.T) {
	original := models.ModuleFederationMetadata{
		Scope:       "dashboard",
		Module:      "widget-overview",
		ImportName:  "OverviewWidget",
		FeatureFlag: "feature.widget.enabled",
		Defaults: models.BaseWidgetDimensions{
			Width:     6,
			Height:    4,
			MaxHeight: 8,
			MinHeight: 2,
		},
		Config: models.WidgetConfiguration{
			Title: "Overview Widget",
			Icon:  models.BellIcon,
			HeaderLink: models.WidgetHeaderLink{
				Title:       "Copy",
				Href:        "https://example.com/overview",
				FeatureFlag: "CopyFlag",
			},
			Permissions: []models.WidgetPermission{
				{
					Method: models.WidgetPermissionMethods("GET"),
					Apps:   []string{"app1", "app2"},
					Args:   []any{"arg1", float64(42), true},
				},
			},
		},
	}

	copy, err := deepCopyJSON(original)

	assert.NoError(t, err)
	assert.Equal(t, original, copy)
	assert.NotSame(t, &original, &copy)
}

func TestDeepCopyJSONEmptyMetadata(t *testing.T) {
	original := models.ModuleFederationMetadata{}
	copy, err := deepCopyJSON(original)

	assert.NoError(t, err)
	assert.Equal(t, original, copy)
}
