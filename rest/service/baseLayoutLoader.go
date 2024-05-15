package service

import (
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
		logrus.Errorln("error discovering widget dashboard yaml files")
	}
	// yml extensions
	modulesFiles2, err := filepath.Glob(templatesWD + "widget-dashboard-defaults/*.yml")
	if err != nil {
		logrus.Errorln("error discovering widget dashboard yml files")
	}
	modulesFiles = append(modulesFiles, modulesFiles2...)

	if len(modulesFiles) == 0 {
		logrus.Errorln("no widget dashboard files found")
	}

	for _, file := range modulesFiles {
		var template models.BaseDashboardTemplate
		var jsonTemplate map[string]interface{}
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			logrus.Errorln("error reading widget dashboard file" + file)
		}
		err = yaml.Unmarshal(yamlFile, &template)
		if err != nil {
			logrus.Errorln("error Unmarshal widget dashboard file" + file)
			logrus.Errorln(err)
			break
		}
		dashboardType := models.AvailableTemplates(template.Name)
		err = dashboardType.IsValid()
		if err != nil {
			logrus.Errorln("unknown dashboard type: " + dashboardType)
			logrus.Errorln(err)
			break
		}

		err = template.TemplateConfig.IsValid()

		if err != nil {
			logrus.Errorln("invalid template config in file: " + file + "\n" + err.Error())
			break
		}
		err = yaml.Unmarshal(yamlFile, &jsonTemplate)

		if err != nil {
			logrus.Errorln("invalid template config in file: " + file + "\n" + err.Error())
			break
		}

		baseLayouts[dashboardType] = template
		foundTemplateDefaults[dashboardType] = true
	}

	for template, found := range foundTemplateDefaults {
		if !found {
			logrus.Errorln("missing default template for " + template.String())
		}
	}

	BaseTemplates = baseLayouts
}
