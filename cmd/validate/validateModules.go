package main

import (
	"fmt"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

func validateModules(cwd string) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://./static/modulesSchema.json")
	modulesFiles, err := filepath.Glob(cwd + "/static/**/**/**/fed-modules.json")
	handleErr(err)

	for _, file := range modulesFiles {
		documentLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", file))
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			panic(err.Error())
		}

		if !result.Valid() {
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
			panic(fmt.Sprintf("The %s is not valid. see errors :\n", file))
		}
	}
}
