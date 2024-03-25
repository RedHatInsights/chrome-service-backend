package util

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"gorm.io/datatypes"
)

const (
	XRHIDENTITY      = "x-rh-identity"
	LAST_VISITED_MAX = 10
	IDENTITY_CTX_KEY = "identity"
	USER_CTX_KEY     = "user"
	GET_ALL_PARAM    = "getAll"   // Used for searching ALL favorited pages
	DEFAULT_PARAM    = "archived" // Used as default value for active favorited pages
	FAVORITE_PARAM   = "favorite"
)

func prepareInitialGridItems(jsonData string) []models.GridItem {
	var items []models.GridItem
	json.Unmarshal([]byte(jsonData), &items)

	return items
}

// TODO: replace these once we have actual base templates
func getLandingPageBaseLayout(x string) string {
	if x == "" {
		x = "1"
	}
	const template = `[
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw1", "x": 0, "y": 0, "static": true},
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw2", "x": 0, "y": 1, "static": true },
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw3", "x": 0, "y": 2, "static": true},
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "MediumWidget#mw1", "x": %[1]s, "y": 2, "static": true},
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "SmallWidget#sw1", "x": %[1]s, "y": 0, "static": true},
		{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "SmallWidget#sw2", "x": %[1]s, "y": 1, "static": true}
	]`
	return fmt.Sprintf(template, x)
}

var landingPageSm = prepareInitialGridItems(getLandingPageBaseLayout("1"))
var landingPageMd = prepareInitialGridItems(getLandingPageBaseLayout("2"))
var landingPageLg = prepareInitialGridItems(getLandingPageBaseLayout("3"))
var landingPageXl = prepareInitialGridItems(getLandingPageBaseLayout("4"))

func convertToJson(items []models.GridItem) datatypes.JSONType[[]models.GridItem] {
	gi := datatypes.NewJSONType(items)
	return gi
}

var (
	ErrNotAuthorized                      = errors.New("not authorized")
	BaseTemplates    models.BaseTemplates = models.BaseTemplates{
		"landingPage": models.BaseDashboardTemplate{
			Name:        "landingPage",
			DisplayName: "Landing Page",
			TemplateConfig: models.TemplateConfig{
				Sm: convertToJson(landingPageSm),
				Md: convertToJson(landingPageMd),
				Lg: convertToJson(landingPageLg),
				Xl: convertToJson(landingPageXl),
			},
		},
	}
)
