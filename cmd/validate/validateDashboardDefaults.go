package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

func validateDashboardDefaults(cwd string) {
	compiler := jsonschema.NewCompiler()

	foundTemplateDefaults := map[models.AvailableTemplates]bool{
		models.LandingPage: false,
	}

	fmt.Println("Validating dashboard defaults", cwd)
	schemaFile, err := os.ReadFile(cwd + "/widget-dashboard-defaults/widget-schema.json")

	if err != nil {
		panic(err)
	}
	err = compiler.AddResource("schema.json", strings.NewReader(string(schemaFile)))

	if err != nil {
		panic(err)
	}

	schema, err := compiler.Compile("schema.json")

	if err != nil {
		panic(err)
	}

	// yaml extensions
	modulesFiles, err := filepath.Glob(cwd + "/widget-dashboard-defaults/*.yaml")
	if err != nil {
		panic(err)
	}
	// yml extensions
	modulesFiles2, err := filepath.Glob(cwd + "/widget-dashboard-defaults/*.yml")
	if err != nil {
		panic(err)
	}
	modulesFiles = append(modulesFiles, modulesFiles2...)

	for _, file := range modulesFiles {
		var yamlData map[string]interface{}
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(yamlFile, &yamlData)
		if err != nil {
			panic(err)
		}

		name := yamlData["name"].(string)
		if foundTemplateDefaults[models.AvailableTemplates(name)] {
			panic("Duplicate dashboard default name: " + name)
		}
		foundTemplateDefaults[models.AvailableTemplates(name)] = true

		err = schema.Validate(yamlData)

		if err != nil {
			fmt.Println("File", file)
			panic(err.Error())
		}
	}

	for template, found := range foundTemplateDefaults {
		if !found {
			panic("Missing dashboard default: " + string(template))
		}
	}
}
