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
	"gorm.io/gorm"
)

var modifiedTemplate1 models.DashboardTemplate
var removableTemplate models.DashboardTemplate
var template1 models.DashboardTemplate
var template2 models.DashboardTemplate

func mockDashboardTemplatesData() {
	identity := models.UserIdentity{
		AccountId: "1",
	}

	emptyIdentity := models.UserIdentity{
		AccountId: "2",
	}

	database.DB.Create(&identity)
	database.DB.Create(&emptyIdentity)

	template1 = models.DashboardTemplate{
		UserIdentityID: identity.ID,
		Default:        true,
		TemplateBase: models.DashboardTemplateBase{
			Name:        models.LandingPage.String(),
			DisplayName: "Template 1",
		},
		TemplateConfig: models.TemplateConfig{
			Sx: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Md: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Lg: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Xl: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
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
			Sx: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Md: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Lg: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
			Xl: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
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
			Sx: []byte(`[{"title":"","i":"foo","x":0,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
			Md: []byte(`[{"title":"","i":"1","x":0,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
			Lg: []byte(`[{"title":"","i":"1","x":0,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
			Xl: []byte(`[{"title":"","i":"1","x":0,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
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

	database.DB.Create(&template1)
	database.DB.Create(&template2)
	database.DB.Create(&modifiedTemplate1)
	database.DB.Create(&removableTemplate)
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
	time := time.Now().UnixNano()
	dbName = fmt.Sprintf("%d-services.db", time)
	config.Get().DbName = dbName

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
				Sx: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Md: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Lg: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Xl: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
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
				Sx: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Md: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Lg: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
				Xl: []byte(`[{"i": "1", "x": 0, "y": 0, "w": 1, "h": 1, "maxH": 4, "minH": 1}]`),
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
				Sx: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
				Md: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
				Lg: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
				Xl: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
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
				Sx: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":0,"h":1,"maxH":4,"minH":1}]`),
				Md: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
				Lg: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
				Xl: []byte(`[{"title":"","i":"1","x":1,"y":0,"w":1,"h":1,"maxH":4,"minH":1}]`),
			},
		}
		_, err := UpdateDashboardTemplate(templateId, userId, template)
		assert.NotNil(t, err)
		assert.Equal(t, `invalid grid item, height "h", width "w", maxHeight "maxH", mixHeight "minH" must be greater than 0`, err.Error())
	})

	t.Run("GetAllBaseTemplates should return all base templates", func(t *testing.T) {
		baseTemplates := GetAllBaseTemplates()
		assert.Equal(t, 1, len(baseTemplates))
		assert.Equal(t, util.BaseTemplates[models.LandingPage].Name, baseTemplates[0].Name)
		assert.Equal(t, util.BaseTemplates[models.LandingPage].DisplayName, baseTemplates[0].DisplayName)
		assert.Equal(t, util.BaseTemplates[models.LandingPage].TemplateConfig, baseTemplates[0].TemplateConfig)
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
			assert.Equal(t, util.BaseTemplates[templateType].DisplayName, templateBase.DisplayName)
			assert.Equal(t, util.BaseTemplates[templateType].TemplateConfig, templateBase.TemplateConfig)
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
}
