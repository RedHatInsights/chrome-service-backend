package util

import (
	"encoding/json"
	"errors"

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
var landingPageBaseLayout = `[
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw1", "x": 0, "y": 0 },
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw2", "x": 0, "y": 1 },
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "LargeWidget#lw3", "x": 0, "y": 2 },
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "MediumWidget#mw1", "x": 4, "y": 2 },
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "SmallWidget#sw1", "x": 4, "y": 0 },
	{ "maxH": 4, "minH": 1, "w": 1, "h": 1, "title": "Widget 1", "i": "SmallWidget#sw2", "x": 4, "y": 1 }
]`

var landingPage = prepareInitialGridItems(landingPageBaseLayout)

func convertToJson(items []models.GridItem) datatypes.JSON {
	json, err := json.Marshal(items)
	if err != nil {
		panic(err)
	}

	return json
}

var (
	ErrNotAuthorized                      = errors.New("not authorized")
	BaseTemplates    models.BaseTemplates = models.BaseTemplates{
		"landingPage": models.BaseDashboardTemplate{
			Name:        "landingPage",
			DisplayName: "Landing Page",
			TemplateConfig: models.TemplateConfig{
				Sx: convertToJson(landingPage),
				Md: convertToJson(landingPage),
				Lg: convertToJson(landingPage),
				Xl: convertToJson(landingPage),
			},
		},
	}
)
