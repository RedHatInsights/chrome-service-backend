package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func LoadBaseLayout() {
	foundTemplateDefaults := map[models.AvailableTemplates]bool{
		models.LandingPage: false,
	}
	baseLayouts := models.BaseTemplates{}
	templatesWD := config.Get().DashboardConfig.TemplatesWD
	// yaml extensions
	modulesFiles, err := filepath.Glob(templatesWD + "widget-dashboard-defaults/*.yaml")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "dashboard_templates",
			"operation":  "discover",
			"extension":  "yaml",
			"error_type": fmt.Sprintf("%T", err),
		}).Error("dashboard template discovery failed")
	}
	// yml extensions
	modulesFiles2, err := filepath.Glob(templatesWD + "widget-dashboard-defaults/*.yml")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component":  "dashboard_templates",
			"operation":  "discover",
			"extension":  "yml",
			"error_type": fmt.Sprintf("%T", err),
		}).Error("dashboard template discovery failed")
	}
	modulesFiles = append(modulesFiles, modulesFiles2...)

	if len(modulesFiles) == 0 {
		logrus.WithFields(logrus.Fields{
			"component": "dashboard_templates",
			"operation": "discover",
		}).Error("no dashboard template files found")
	}

	for _, file := range modulesFiles {
		var template models.BaseDashboardTemplate
		var jsonTemplate map[string]interface{}
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "dashboard_templates",
				"operation":  "read",
				"error_type": fmt.Sprintf("%T", err),
			}).Error("dashboard template file read failed")
			continue
		}
		err = yaml.Unmarshal(yamlFile, &template)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "dashboard_templates",
				"operation":  "decode",
				"error_type": fmt.Sprintf("%T", err),
			}).Error("dashboard template decode failed")
			break
		}
		dashboardType := models.AvailableTemplates(template.Name)
		err = dashboardType.IsValid()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "dashboard_templates",
				"operation":  "validate_type",
				"error_type": fmt.Sprintf("%T", err),
			}).Error("dashboard template type is invalid")
			break
		}

		err = template.TemplateConfig.IsValid()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "dashboard_templates",
				"operation":  "validate_config",
				"error_type": fmt.Sprintf("%T", err),
			}).Error("dashboard template config is invalid")
			break
		}
		err = yaml.Unmarshal(yamlFile, &jsonTemplate)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component":  "dashboard_templates",
				"operation":  "decode_config",
				"error_type": fmt.Sprintf("%T", err),
			}).Error("dashboard template config decode failed")
			break
		}

		baseLayouts[dashboardType] = template
		foundTemplateDefaults[dashboardType] = true
	}

	for template, found := range foundTemplateDefaults {
		if !found {
			logrus.WithFields(logrus.Fields{
				"component": "dashboard_templates",
				"operation": "validate_defaults",
				"template":  template.String(),
			}).Error("dashboard template default is missing")
		}
	}

	BaseTemplates = baseLayouts
}
