package service

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	encodedTemplateString = "eyJjcmVhdGVkQXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsInVwZGF0ZWRBdCI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIiwiZGVsZXRlZEF0IjpudWxsLCJ1c2VySWRlbnRpdHlJRCI6MCwiZGVmYXVsdCI6ZmFsc2UsIlRlbXBsYXRlQmFzZSI6eyJuYW1lIjoidGVzdCIsImRpc3BsYXlOYW1lIjoidGVzdCJ9LCJ0ZW1wbGF0ZUNvbmZpZyI6eyJzbSI6W3sidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6InJoZWwjcmhlbCIsIngiOjAsInkiOjAsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6Im9wZW5zaGlmdCNvcGVuc2hpZnQiLCJ4IjowLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhbnNpYmxlI2Fuc2libGUiLCJ4IjowLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjUsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJleHBsb3JlQ2FwYWJpbGl0aWVzI2V4cGxvcmVDYXBhYmlsaXRpZXMiLCJ4IjowLCJ5IjoyLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjcsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJyZWNlbnRseVZpc2l0ZWQjcmVjZW50bHlWaXNpdGVkIiwieCI6MCwieSI6MCwic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo2LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoiZmF2b3JpdGVTZXJ2aWNlcyNmYXZvcml0ZVNlcnZpY2VzIiwieCI6MCwieSI6Mywic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo0LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoib3BlbnNoaWZ0QWkjb3BlbnNoaWZ0QWkiLCJ4IjowLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJlZGdlI2VkZ2UiLCJ4IjowLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhY3MjYWNzIiwieCI6MCwieSI6Mywic3RhdGljIjpmYWxzZX1dLCJtZCI6W3sidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6InJoZWwjcmhlbCIsIngiOjAsInkiOjAsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6Im9wZW5zaGlmdCNvcGVuc2hpZnQiLCJ4IjoxLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhbnNpYmxlI2Fuc2libGUiLCJ4IjoxLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjIsImgiOjUsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJleHBsb3JlQ2FwYWJpbGl0aWVzI2V4cGxvcmVDYXBhYmlsaXRpZXMiLCJ4IjowLCJ5IjoyLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjcsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJyZWNlbnRseVZpc2l0ZWQjcmVjZW50bHlWaXNpdGVkIiwieCI6MSwieSI6MCwic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo2LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoiZmF2b3JpdGVTZXJ2aWNlcyNmYXZvcml0ZVNlcnZpY2VzIiwieCI6MSwieSI6Mywic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo0LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoib3BlbnNoaWZ0QWkjb3BlbnNoaWZ0QWkiLCJ4IjowLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJlZGdlI2VkZ2UiLCJ4IjoxLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhY3MjYWNzIiwieCI6MSwieSI6Mywic3RhdGljIjpmYWxzZX1dLCJsZyI6W3sidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6InJoZWwjcmhlbCIsIngiOjAsInkiOjAsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6Im9wZW5zaGlmdCNvcGVuc2hpZnQiLCJ4IjoxLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhbnNpYmxlI2Fuc2libGUiLCJ4IjoyLCJ5IjoxMywic3RhdGljIjpmYWxzZX0seyJ3IjozLCJoIjo1LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoiZXhwbG9yZUNhcGFiaWxpdGllcyNleHBsb3JlQ2FwYWJpbGl0aWVzIiwieCI6MCwieSI6OCwic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo4LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoicmVjZW50bHlWaXNpdGVkI3JlY2VudGx5VmlzaXRlZCIsIngiOjIsInkiOjAsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NiwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6ImZhdm9yaXRlU2VydmljZXMjZmF2b3JpdGVTZXJ2aWNlcyIsIngiOjAsInkiOjEzLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJvcGVuc2hpZnRBaSNvcGVuc2hpZnRBaSIsIngiOjAsInkiOjQsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6ImVkZ2UjZWRnZSIsIngiOjEsInkiOjQsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6ImFjcyNhY3MiLCJ4IjoxLCJ5IjoxMywic3RhdGljIjpmYWxzZX1dLCJ4bCI6W3sidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6InJoZWwjcmhlbCIsIngiOjAsInkiOjAsInN0YXRpYyI6ZmFsc2V9LHsidyI6MSwiaCI6NCwibWF4SCI6MTAsIm1pbkgiOjEsInRpdGxlIjoiIiwiaSI6Im9wZW5zaGlmdCNvcGVuc2hpZnQiLCJ4IjoxLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhbnNpYmxlI2Fuc2libGUiLCJ4IjoyLCJ5IjowLCJzdGF0aWMiOmZhbHNlfSx7InciOjMsImgiOjUsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJleHBsb3JlQ2FwYWJpbGl0aWVzI2V4cGxvcmVDYXBhYmlsaXRpZXMiLCJ4IjowLCJ5IjoyLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjcsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJyZWNlbnRseVZpc2l0ZWQjcmVjZW50bHlWaXNpdGVkIiwieCI6MywieSI6MCwic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo2LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoiZmF2b3JpdGVTZXJ2aWNlcyNmYXZvcml0ZVNlcnZpY2VzIiwieCI6MywieSI6Mywic3RhdGljIjpmYWxzZX0seyJ3IjoxLCJoIjo0LCJtYXhIIjoxMCwibWluSCI6MSwidGl0bGUiOiIiLCJpIjoib3BlbnNoaWZ0QWkjb3BlbnNoaWZ0QWkiLCJ4IjowLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJlZGdlI2VkZ2UiLCJ4IjoxLCJ5IjozLCJzdGF0aWMiOmZhbHNlfSx7InciOjEsImgiOjQsIm1heEgiOjEwLCJtaW5IIjoxLCJ0aXRsZSI6IiIsImkiOiJhY3MjYWNzIiwieCI6MiwieSI6Mywic3RhdGljIjpmYWxzZX1dfX0K"
)

var modifiedTemplate1 models.DashboardTemplate
var removableTemplate models.DashboardTemplate
var template1 models.DashboardTemplate
var template2 models.DashboardTemplate
var encodingTemplate models.DashboardTemplate

func getMockItems() datatypes.JSONType[[]models.GridItem] {
	return datatypes.NewJSONType([]models.GridItem{
		{
			ID: "1",
			X:  0,
			Y:  0,
			BaseWidgetDimensions: models.BaseWidgetDimensions{
				Width:     1,
				Height:    1,
				MaxHeight: 4,
				MinHeight: 1,
			},
		},
	})
}

func mockDashboardTemplatesData() {
	identity := models.UserIdentity{
		AccountId: "1",
	}

	emptyIdentity := models.UserIdentity{
		AccountId: "2",
	}

	encodeIdentity := models.UserIdentity{
		AccountId: "3",
	}

	database.DB.Create(&identity)
	database.DB.Create(&emptyIdentity)
	database.DB.Create(&encodeIdentity)

	template1 = models.DashboardTemplate{
		UserIdentityID: identity.ID,
		Default:        true,
		TemplateBase: models.DashboardTemplateBase{
			Name:        models.LandingPage.String(),
			DisplayName: "Template 1",
		},
		TemplateConfig: models.TemplateConfig{
			Sm: getMockItems(),
			Md: getMockItems(),
			Lg: getMockItems(),
			Xl: getMockItems(),
		},
	}

	template2 = models.DashboardTemplate{
		UserIdentityID: identity.ID,
		Default:        false,
		TemplateBase: models.DashboardTemplateBase{
			Name:        "fakeTemplate",
			DisplayName: "Template 2",
		},
		TemplateConfig: models.TemplateConfig{
			Sm: getMockItems(),
			Md: getMockItems(),
			Lg: getMockItems(),
			Xl: getMockItems(),
		},
	}

	modifiedTemplate1 = models.DashboardTemplate{
		UserIdentityID: identity.ID,
		Default:        false,
		TemplateBase: models.DashboardTemplateBase{
			Name:        models.LandingPage.String(),
			DisplayName: "Modified Template 1",
		},
		TemplateConfig: models.TemplateConfig{
			Sm: datatypes.NewJSONType([]models.GridItem{{
				ID: "foo",
				X:  0,
				Y:  0,
				BaseWidgetDimensions: models.BaseWidgetDimensions{
					Width:     1,
					Height:    1,
					MaxHeight: 4,
					MinHeight: 1,
				},
			}}),
			Md: datatypes.NewJSONType([]models.GridItem{
				{
					ID: "1",
					X:  0,
					Y:  0,
					BaseWidgetDimensions: models.BaseWidgetDimensions{
						Width:     1,
						Height:    1,
						MaxHeight: 4,
						MinHeight: 1,
					},
				},
			}),
			Lg: getMockItems(),
			Xl: getMockItems(),
		},
	}

	removableTemplate = models.DashboardTemplate{
		UserIdentityID: identity.ID,
		Default:        false,
		TemplateBase: models.DashboardTemplateBase{
			Name:        "removableTemplate",
			DisplayName: "Removable Template",
		},
		TemplateConfig: template1.TemplateConfig,
	}

	encodingTemplate = models.DashboardTemplate{
		UserIdentityID: encodeIdentity.ID,
		Default:        false,
		TemplateBase: models.DashboardTemplateBase{
			Name:        "test",
			DisplayName: "test",
		},
		TemplateConfig: BaseTemplates["landingPage"].TemplateConfig,
	}

	database.DB.Create(&template1)
	database.DB.Create(&template2)
	database.DB.Create(&modifiedTemplate1)
	database.DB.Create(&removableTemplate)
	database.DB.Create(&encodingTemplate)
}

func TestMain(m *testing.M) {
	setUp()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

var dbName string

func setUp() {
	cfg := config.Get()
	database.Init()
	cfg.Test = true
	// This is critical for the dashboard template loader CWD
	cfg.DashboardConfig.TemplatesWD = "../../"
	time := time.Now().UnixNano()
	dbName = fmt.Sprintf("%d-services.db", time)
	config.Get().DbName = dbName
	LoadBaseLayout()

	database.Init()
	err := database.DB.AutoMigrate(&models.DashboardTemplate{}, &models.UserIdentity{})
	if err != nil {
		panic(err)
	}
}

func tearDown() {
	os.Remove(dbName)
}

func TestGetAllUserDashboardTemplates(t *testing.T) {
	mockDashboardTemplatesData()
	t.Run("Test Get All User Dashboard Templates", func(t *testing.T) {
		userId := uint(1)
		userDashboardTemplates, err := GetAllUserDashboardTemplates(userId)
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 4, len(userDashboardTemplates))
		assert.Equal(t, models.LandingPage.String(), userDashboardTemplates[0].TemplateBase.Name)
	})

	t.Run("Should return empty array if no templates found", func(t *testing.T) {
		userId := uint(2)
		userDashboardTemplates, err := GetAllUserDashboardTemplates(userId)
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 0, len(userDashboardTemplates))
	})

	t.Run("Should return dashboard templates  with landingPage base template", func(t *testing.T) {
		userId := uint(1)
		userDashboardTemplates, err := GetUserDashboardTemplate(userId, models.LandingPage)
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 2, len(userDashboardTemplates))
		assert.Equal(t, models.LandingPage.String(), userDashboardTemplates[0].TemplateBase.Name)
	})

	t.Run("Should create new dashboard template with landingPage base template if user does not have personalized landingPage template", func(t *testing.T) {
		userId := uint(2)
		userDashboardTemplates, err := GetUserDashboardTemplate(userId, models.LandingPage)
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 1, len(userDashboardTemplates))
		assert.Equal(t, models.LandingPage.String(), userDashboardTemplates[0].TemplateBase.Name)
	})

	t.Run("GetDashboardTemplate should return all templates if dashboard is empty", func(t *testing.T) {
		userId := uint(1)
		userDashboardTemplates, err := GetDashboardTemplate(userId, models.AvailableTemplates(""))
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 4, len(userDashboardTemplates))
		assert.Equal(t, models.LandingPage.String(), userDashboardTemplates[0].TemplateBase.Name)
		assert.Equal(t, "fakeTemplate", userDashboardTemplates[1].TemplateBase.Name)
	})

	t.Run("GetDashboardTemplate should return only landingPage dashboard templates if dashboard is not landingPage", func(t *testing.T) {
		userId := uint(1)
		userDashboardTemplates, err := GetDashboardTemplate(userId, models.LandingPage)
		assert.Nil(t, err)
		assert.NotNil(t, userDashboardTemplates)
		assert.Equal(t, 2, len(userDashboardTemplates))
		assert.Equal(t, models.LandingPage.String(), userDashboardTemplates[0].TemplateBase.Name)
	})

	t.Run("UpdateDashboardTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(100)
		template := models.DashboardTemplate{
			TemplateBase: models.DashboardTemplateBase{
				Name:        models.LandingPage.String(),
				DisplayName: "Template 1",
			},
			TemplateConfig: models.TemplateConfig{
				Sm: getMockItems(),
				Md: getMockItems(),
				Lg: getMockItems(),
				Xl: getMockItems(),
			},
		}
		_, err := UpdateDashboardTemplate(templateId, userId, template)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("UpdateDashboardTemplate should return ErrNotAuthorized if user does not own the template", func(t *testing.T) {
		userId := uint(2)
		templateId := uint(1)
		template := models.DashboardTemplate{
			TemplateBase: models.DashboardTemplateBase{
				Name:        models.LandingPage.String(),
				DisplayName: "Template 1",
			},
			TemplateConfig: models.TemplateConfig{
				Sm: getMockItems(),
				Md: getMockItems(),
				Lg: getMockItems(),
				Xl: getMockItems(),
			},
		}
		_, err := UpdateDashboardTemplate(templateId, userId, template)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, util.ErrNotAuthorized))
	})

	t.Run("UpdateDashboardTemplate should only update the template config, rest of attributes should remain unchanged", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(1)
		template := models.DashboardTemplate{
			TemplateBase: models.DashboardTemplateBase{
				Name:        models.LandingPage.String(),
				DisplayName: "Foo bar",
			},
			TemplateConfig: models.TemplateConfig{
				Sm: getMockItems(),
				Md: getMockItems(),
				Lg: getMockItems(),
				Xl: getMockItems(),
			},
		}
		updatedTemplate, err := UpdateDashboardTemplate(templateId, userId, template)
		assert.Nil(t, err)
		assert.NotNil(t, updatedTemplate)
		assert.Equal(t, template.TemplateConfig, updatedTemplate.TemplateConfig)
		assert.Equal(t, models.LandingPage.String(), updatedTemplate.TemplateBase.Name)
		assert.Equal(t, "Template 1", updatedTemplate.TemplateBase.DisplayName)
		assert.Equal(t, uint(1), updatedTemplate.UserIdentityID)
		assert.Equal(t, true, updatedTemplate.Default)
	})

	t.Run("UpdateDashboardTemplate should return an error if template is not valid", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(1)
		template := models.DashboardTemplate{
			TemplateBase: models.DashboardTemplateBase{
				Name:        models.LandingPage.String(),
				DisplayName: "Foo bar",
			},
			TemplateConfig: models.TemplateConfig{
				Sm: datatypes.NewJSONType([]models.GridItem{
					{
						ID: "1",
						X:  1,
						Y:  0,
						BaseWidgetDimensions: models.BaseWidgetDimensions{
							Width:     0,
							Height:    1,
							MaxHeight: 4,
							MinHeight: 1,
						},
					},
				}),
				Md: getMockItems(),
				Lg: getMockItems(),
				Xl: getMockItems(),
			},
		}
		_, err := UpdateDashboardTemplate(templateId, userId, template)
		assert.NotNil(t, err)
		assert.Equal(t, `invalid grid item, height "h", width "w", maxHeight "maxH", mixHeight "minH" must be greater than 0`, err.Error())
	})

	t.Run("GetAllBaseTemplates should return all base templates", func(t *testing.T) {
		baseTemplates := GetAllBaseTemplates()
		assert.Equal(t, 1, len(baseTemplates))
		assert.Equal(t, BaseTemplates[models.LandingPage].Name, baseTemplates[0].Name)
		assert.Equal(t, BaseTemplates[models.LandingPage].DisplayName, baseTemplates[0].DisplayName)
		assert.Equal(t, BaseTemplates[models.LandingPage].TemplateConfig, baseTemplates[0].TemplateConfig)
	})

	t.Run("GetDashboardTemplateBase should return error if template type does not exist", func(t *testing.T) {
		_, err := GetDashboardTemplateBase("fakeTemplate")
		assert.NotNil(t, err)
		assert.Equal(t, "invalid dashboard template. Expected one of landingPage, got fakeTemplate", err.Error())
	})

	t.Run("GetDashboardTemplateBase should return template for existing template type", func(t *testing.T) {
		existingTemplates := []models.AvailableTemplates{models.LandingPage}
		for _, templateType := range existingTemplates {
			templateBase, err := GetDashboardTemplateBase(templateType)
			assert.Nil(t, err)
			assert.NotNil(t, templateBase)
			assert.Equal(t, templateType.String(), templateBase.Name)
			assert.Equal(t, BaseTemplates[templateType].DisplayName, templateBase.DisplayName)
			assert.Equal(t, BaseTemplates[templateType].TemplateConfig, templateBase.TemplateConfig)
		}
	})

	t.Run("CopyDashboardTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(99999)
		_, err := CopyDashboardTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("CopyDashboardTemplate should copy the template and return the new template with distinct ID", func(t *testing.T) {
		var templateOriginal models.DashboardTemplate
		templateId := modifiedTemplate1.ID
		database.DB.Find(&templateOriginal, modifiedTemplate1.ID)
		userId := uint(1)
		template, err := CopyDashboardTemplate(userId, templateId)
		assert.Nil(t, err)
		assert.NotNil(t, template)
		assert.NotEqual(t, templateOriginal.ID, template.ID)
		assert.Equal(t, templateOriginal.UserIdentityID, template.UserIdentityID)
		assert.Equal(t, templateOriginal.Default, template.Default)
		assert.Equal(t, templateOriginal.TemplateBase.DisplayName, template.TemplateBase.DisplayName)
		assert.Equal(t, templateOriginal.TemplateBase.Name, template.TemplateBase.Name)
		assert.Equal(t, templateOriginal.TemplateConfig, template.TemplateConfig)
	})

	t.Run("DeleteTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(99999)
		err := DeleteTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("DeleteTemplate should return ErrNotAuthorized if user does not own the template", func(t *testing.T) {
		userId := uint(2)
		templateId := uint(1)
		err := DeleteTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, util.ErrNotAuthorized))
	})

	t.Run("DeleteTemplate should delete the template", func(t *testing.T) {
		userId := uint(1)
		templateId := removableTemplate.ID
		err := DeleteTemplate(userId, templateId)
		assert.Nil(t, err)
		var template models.DashboardTemplate
		database.DB.Find(&template, templateId)
		assert.Equal(t, uint(0), template.ID)
	})

	t.Run("ChangeDefaultTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(99999)
		_, err := ChangeDefaultTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("ChangeDefaultTemplate should return ErrNotAuthorized if user does not own the template", func(t *testing.T) {
		userId := uint(2)
		templateId := uint(1)
		_, err := ChangeDefaultTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, util.ErrNotAuthorized))
	})

	t.Run("ChangeDefaultTemplate should change the landingPage default template to template2", func(t *testing.T) {
		var userLandingTemplates []models.DashboardTemplate
		userId := uint(1)
		templateId := modifiedTemplate1.ID
		template, err := ChangeDefaultTemplate(userId, templateId)
		assert.Nil(t, err)
		assert.NotNil(t, template)
		assert.Equal(t, true, template.Default)
		database.DB.Where("user_identity_id = ? AND name = ?", userId, models.LandingPage).Find(&userLandingTemplates)
		for _, template := range userLandingTemplates {
			shouldBeDefault := template.ID == templateId
			assert.Equal(t, shouldBeDefault, template.Default)
		}
	})

	t.Run("ForkBaseTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		_, err := ForkBaseTemplate(userId, "fakeTemplate")
		assert.NotNil(t, err)
		assert.Equal(t, "invalid dashboard template. Expected one of landingPage, got fakeTemplate", err.Error())
	})

	t.Run("ForkBaseTemplate should create a new template with the base template", func(t *testing.T) {
		userId := uint(1)
		template, err := ForkBaseTemplate(userId, "landingPage")
		assert.Nil(t, err)
		assert.NotNil(t, template)
		assert.Equal(t, models.LandingPage.String(), template.TemplateBase.Name)
	})

	t.Run("EncodeDashboardTemplate should return not found error if template does not exist", func(t *testing.T) {
		userId := uint(1)
		templateId := uint(99999)
		_, err := EncodeDashboardTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("EncodeDashboardTemplate should return ErrNotAuthorized if user does not own the template", func(t *testing.T) {
		userId := uint(2)
		templateId := uint(1)
		_, err := EncodeDashboardTemplate(userId, templateId)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, util.ErrNotAuthorized))
	})

	t.Run("EncodeDashboardTemplate should return the encoded template", func(t *testing.T) {
		userId := uint(3)
		templateId := encodingTemplate.ID
		encoded, err := EncodeDashboardTemplate(userId, templateId)
		assert.Nil(t, err)
		assert.NotNil(t, encoded)
		assert.Equal(t, encodedTemplateString, encoded)
	})

	t.Run("Should decode dashboard template", func(t *testing.T) {
		decoded, err := DecodeDashboardTemplate(encodedTemplateString)
		assert.Nil(t, err)
		assert.NotNil(t, decoded)
		assert.Equal(t, encodingTemplate.TemplateBase.Name, decoded.TemplateBase.Name)
		assert.Equal(t, encodingTemplate.TemplateBase.DisplayName, decoded.TemplateBase.DisplayName)
		assert.Equal(t, encodingTemplate.TemplateConfig, decoded.TemplateConfig)
	})
}
