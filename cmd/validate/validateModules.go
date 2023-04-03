package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/gookit/validate"
)

type route struct {
	Pathname  string `json:"pathname" validate:"required"`
	Exact     bool   `json:"exact"`
	IsFedramp bool   `json:"isFedramp"`
}

type routeModule struct {
	Id     string  `json:"id" validate:"required"`
	Module string  `json:"module" validate:"required"`
	Routes []route `json:"routes" validate:"required"`
}

type analyticsConfig struct {
	APIKey *string `json:"APIkey" validate:"minLen:1"`
}

type moduleItem struct {
	DefaultDocumentTitle string          `json:"defaultDocumentTitle,omitempty"`
	ManifestLocation     string          `json:"manifestLocation" validate:"required"`
	IsFedramp            bool            `json:"isFedramp"`
	Modules              []routeModule   `json:"modules"`
	Analytics            analyticsConfig `json:"analytics,omitempty"`
}

func validateModules(cwd string) error {
	// get the all fed-modules.json files
	modulesFiles, err := filepath.Glob(cwd + "/static/**/**/**/fed-modules.json")
	for _, filePath := range modulesFiles {
		file, err := ioutil.ReadFile(filePath)
		handleErr(err)
		var m map[string]moduleItem
		err = json.Unmarshal([]byte(file), &m)
		handleErr(err)
		// iterate over values
		for moduleId, v := range m {
			res := validate.Struct(&v)
			if !res.Validate() {
				handleValidationError(res, moduleId)

			}
		}

	}
	return err
}
