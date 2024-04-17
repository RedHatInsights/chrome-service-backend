package service

import (
	"reflect"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var landingPageSm = getLandingPageBaseLayout(1)
var landingPageMd = getLandingPageBaseLayout(2)
var landingPageLg = getLandingPageBaseLayout(3)
var landingPageXl = getLandingPageBaseLayout(4)

func convertToJson(items []models.GridItem) datatypes.JSONType[[]models.GridItem] {
	gi := datatypes.NewJSONType(items)
	return gi
}

var (
	BaseTemplates models.BaseTemplates = models.BaseTemplates{
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
	WidgetMapping models.WidgetModuleFederationMapping = models.WidgetModuleFederationMapping{
		models.ExploreCapabilities: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./ExploreCapabilities",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 3, 5, 6, 1),

			Config: models.WidgetConfiguration{
				Icon:  models.RocketIcon,
				Title: "Explore capabilities",
			},
		},
		models.Edge: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./EdgeWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 4, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.EdgeIcon,
				Title: "Edge Management",
			},
		},
		models.Ansible: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./AnsibleWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 4, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.AnsibleIcon,
				Title: "Ansible Automation Platform",
			},
		},
		models.Rhel: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./RhelWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 4, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.RhelIcon,
				Title: "Red Hat Enterprise Linux",
			},
		},
		models.OpenShift: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./OpenShiftWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 4, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.OpenShiftIcon,
				Title: "Red Hat OpenShift",
			},
		},
		models.Quay: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./QuayWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 3, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.QuayIcon,
				Title: "Quay.io",
			},
		},
		models.Acs: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./AcsWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 4, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.ACSIcon,
				Title: "Advanced Cluster Security",
			},
		},
		models.RecentlyVisited: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./RecentlyVisited",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 3, 4, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.HistoryIcon,
				Title: "Recently visited",
			},
		},
		models.OpenShiftAi: models.ModuleFederationMetadata{
			Scope:    "landing",
			Module:   "./OpenShiftAiWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 3, 3, 1),
			Config: models.WidgetConfiguration{
				Icon:  models.OpenShiftAiIcon,
				Title: "Red Hat OpenShift AI",
			},
		},
		models.FavoriteServices: models.ModuleFederationMetadata{
			Scope:    "chrome",
			Module:   "./DashboardFavorites",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 2, 4, 6, 1),
			Config: models.WidgetConfiguration{
				HeaderLink: models.WidgetHeaderLink{
					Title: "View all services",
					Href:  "/allservices",
				},
				Icon:  models.StarIcon,
				Title: "My favorite services",
			},
		},
		models.NotificationsEvents: models.ModuleFederationMetadata{
			Scope:    "notifications",
			Module:   "./DashboardWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 1, 3, 4, 1),
			Config: models.WidgetConfiguration{
				HeaderLink: models.WidgetHeaderLink{
					Title: "View event log",
					Href:  "/settings/notifications/eventlog",
				},
				Icon:  models.BellIcon,
				Title: "Events",
				Permissions: []models.WidgetPermission{
					models.WidgetPermission{
						Method: models.OrgAdmin,
					},
				},
			},
		},
		models.LearningResources: models.ModuleFederationMetadata{
			Scope:    "learningResources",
			Module:   "./BookmarkedLearningResourcesWidget",
			Defaults: models.BaseWidgetDimensions.InitDimensions(models.BaseWidgetDimensions{}, 2, 4, 4, 1),
			Config: models.WidgetConfiguration{
				HeaderLink: models.WidgetHeaderLink{
					Title: "View all learning resources",
					Href:  "/settings/learning-resources#documentation",
				},
				Icon:  models.OutlinedBookmarkIcon,
				Title: "Bookmarked learning resources",
			},
		},
	}
)

func ForkBaseTemplate(userId uint, dashboard models.AvailableTemplates) (models.DashboardTemplate, error) {
	err := dashboard.IsValid()
	if err != nil {
		return models.DashboardTemplate{}, err
	}

	baseTemplate := BaseTemplates[dashboard]

	templateBase := models.DashboardTemplateBase{
		Name:        dashboard.String(),
		DisplayName: BaseTemplates[dashboard].DisplayName,
	}

	dashboardTemplate := models.DashboardTemplate{
		UserIdentityID: userId,
		Default:        true,
		TemplateBase:   templateBase,
		TemplateConfig: baseTemplate.TemplateConfig,
	}

	result := database.DB.Create(&dashboardTemplate)

	ChangeDefaultTemplate(userId, dashboardTemplate.ID)

	return dashboardTemplate, result.Error
}

func GetAllUserDashboardTemplates(userId uint) ([]models.DashboardTemplate, error) {
	var userDashboardTemplates []models.DashboardTemplate

	result := database.DB.Where("user_identity_id = ?", userId).Find(&userDashboardTemplates)

	return userDashboardTemplates, result.Error
}

func GetUserDashboardTemplate(userId uint, dashboard models.AvailableTemplates) ([]models.DashboardTemplate, error) {
	var userDashboardTemplates []models.DashboardTemplate

	result := database.DB.Where("user_identity_id = ? AND name = ?", userId, dashboard).Find(&userDashboardTemplates)

	if result.RowsAffected == 0 {
		if result.RowsAffected == 0 && result.Error != nil {
			return userDashboardTemplates, result.Error
		}

		dashboardTemplate, err := ForkBaseTemplate(userId, dashboard)

		if err != nil {
			return nil, err
		}

		userDashboardTemplates = append(userDashboardTemplates, dashboardTemplate)
	}

	return userDashboardTemplates, result.Error
}

func GetDashboardTemplate(userId uint, dashboard models.AvailableTemplates) ([]models.DashboardTemplate, error) {
	var userDashboardTemplates []models.DashboardTemplate
	var err error
	if dashboard.String() == "" {
		userDashboardTemplates, err = GetAllUserDashboardTemplates(userId)
	} else {
		userDashboardTemplates, err = GetUserDashboardTemplate(userId, dashboard)
	}

	return userDashboardTemplates, err
}

func UpdateDashboardTemplate(templateId uint, userId uint, dashboardTemplate models.DashboardTemplate) (models.DashboardTemplate, error) {
	var userDashboardTemplate models.DashboardTemplate
	var err error

	result := database.DB.Find(&userDashboardTemplate, templateId)
	if result.RowsAffected == 0 || result.Error != nil {
		return userDashboardTemplate, gorm.ErrRecordNotFound
	}

	if userDashboardTemplate.UserIdentityID != userId {
		return userDashboardTemplate, util.ErrNotAuthorized
	}

	configs := reflect.ValueOf(dashboardTemplate.TemplateConfig)
	typeOfS := configs.Type()

	for i := 0; i < configs.NumField(); i++ {
		dgi := configs.Field(i).Interface().(datatypes.JSONType[[]models.GridItem])
		items := dgi.Data()
		layoutSize := typeOfS.Field(i).Tag.Get("json")
		for _, gi := range items {
			// initialize coordinates if they do not exist
			if gi.Y == 0 {
				gi.Y = 0
			}
			if gi.X == 0 {
				gi.X = 0
			}

			err = gi.IsValid(models.GridSizes(layoutSize))
			if err != nil {
				return userDashboardTemplate, err
			}
		}

		if len(items) > 0 {
			// replace only non empty items, not the whole config
			dashboardTemplate.TemplateConfig.SetLayoutSizeItems(typeOfS.Field(i).Name, items)
		}
	}
	err = dashboardTemplate.TemplateConfig.IsValid()
	if err != nil {
		return userDashboardTemplate, err
	}

	// Update only the templates, no other fields are allowed to be updated
	database.DB.Model(&userDashboardTemplate).Updates(models.DashboardTemplate{
		TemplateConfig: dashboardTemplate.TemplateConfig,
	})

	return userDashboardTemplate, err
}

func GetAllBaseTemplates() []models.BaseDashboardTemplate {
	var templates []models.BaseDashboardTemplate
	for _, template := range BaseTemplates {
		templates = append(templates, template)
	}

	return templates
}

func GetDashboardTemplateBase(dashboard models.AvailableTemplates) (models.BaseDashboardTemplate, error) {
	var baseTemplate models.BaseDashboardTemplate

	err := dashboard.IsValid()

	if err != nil {
		return baseTemplate, err
	}

	baseTemplate = BaseTemplates[dashboard]

	return baseTemplate, err
}

func CopyDashboardTemplate(accountId uint, dashboardTemplateId uint) (models.DashboardTemplate, error) {
	var dashboardTemplate models.DashboardTemplate

	result := database.DB.Find(&dashboardTemplate, dashboardTemplateId)
	if result.RowsAffected == 0 || result.Error != nil {
		return dashboardTemplate, gorm.ErrRecordNotFound
	}

	newDashboardTemplate := models.DashboardTemplate{
		UserIdentityID: accountId,
		TemplateBase:   dashboardTemplate.TemplateBase,
		TemplateConfig: dashboardTemplate.TemplateConfig,
	}

	result = database.DB.Create(&newDashboardTemplate)

	return newDashboardTemplate, result.Error
}

func DeleteTemplate(accountId uint, dashboardTemplateId uint) error {
	var dashboardTemplate models.DashboardTemplate

	result := database.DB.Find(&dashboardTemplate, dashboardTemplateId)
	if result.RowsAffected == 0 || result.Error != nil {
		return gorm.ErrRecordNotFound
	}

	if dashboardTemplate.UserIdentityID != accountId {
		return util.ErrNotAuthorized
	}

	database.DB.Delete(&dashboardTemplate)

	return result.Error
}

func ChangeDefaultTemplate(accountId uint, dashboardId uint) (models.DashboardTemplate, error) {
	var dashboardTemplate models.DashboardTemplate

	result := database.DB.Find(&dashboardTemplate, dashboardId)
	if result.RowsAffected == 0 || result.Error != nil {
		return dashboardTemplate, gorm.ErrRecordNotFound
	}

	if dashboardTemplate.UserIdentityID != accountId {
		return dashboardTemplate, util.ErrNotAuthorized
	}

	dashboardType := dashboardTemplate.TemplateBase.Name

	result = database.DB.Model(models.DashboardTemplate{}).Where("user_identity_id = ? AND name = ?", accountId, dashboardType).Update("default", false)

	if result.Error != nil {
		return dashboardTemplate, result.Error
	}

	result = database.DB.Model(&dashboardTemplate).Updates(models.DashboardTemplate{
		Default: true,
	})

	return dashboardTemplate, result.Error
}

// TODO: replace these once we have actual base templates
func getLandingPageBaseLayout(x int) []models.GridItem {
	if x == 0 {
		x = 1
	}

	baseGridItems := []models.GridItem{
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.Rhel].Defaults,
			ID:                   "rhel#rhel",
			X:                    0,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.OpenShift].Defaults,
			ID:                   "openshift#openshift",
			X:                    1,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.Ansible].Defaults,
			ID:                   "ansible#ansible",
			X:                    2,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.Acs].Defaults,
			ID:                   "acs#acs",
			X:                    3,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.ExploreCapabilities].Defaults,
			ID:                   "exploreCapabilities#exploreCapabilities",
			X:                    0,
			Y:                    2,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.RecentlyVisited].Defaults,
			ID:                   "recentlyVisited#recentlyVisited",
			X:                    3,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.Quay].Defaults,
			ID:                   "quay#quay",
			X:                    3,
			Y:                    0,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.FavoriteServices].Defaults,
			ID:                   "favoriteServices#favoriteServices",
			X:                    0,
			Y:                    3,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.Edge].Defaults,
			ID:                   "edge#edge",
			X:                    2,
			Y:                    3,
		},
		models.GridItem{
			BaseWidgetDimensions: WidgetMapping[models.OpenShiftAi].Defaults,
			ID:                   "openshiftAi#openshiftAi",
			X:                    4,
			Y:                    2,
		},
	}

	// adjust for maximum x value
	for i, item := range baseGridItems {
		if item.X > x {
			baseGridItems[i].X = x
		}

		if item.BaseWidgetDimensions.Width > x {
			baseGridItems[i].BaseWidgetDimensions.Width = x
		}
	}

	return baseGridItems
}
func EncodeDashboardTemplate(accountId uint, templateId uint) (string, error) {
	var dashboardTemplate models.DashboardTemplate

	result := database.DB.Find(&dashboardTemplate, templateId)
	if result.RowsAffected == 0 || result.Error != nil {
		return "", gorm.ErrRecordNotFound
	}

	if dashboardTemplate.UserIdentityID != accountId {
		return "", util.ErrNotAuthorized
	}

	encoded, err := dashboardTemplate.EncodeBase64()

	return encoded, err
}

func DecodeDashboardTemplate(encoded string) (models.DashboardTemplate, error) {
	dashboardTemplate, err := models.DecodeDashboardBase64(encoded)

	return dashboardTemplate, err
}
