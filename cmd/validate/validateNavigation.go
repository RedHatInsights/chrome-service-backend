package main

import (
	"fmt"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"github.com/xeipuuv/gojsonschema"
)

var duplicateCounter = make(map[string]int)

func validateNavigation(cwd string) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://./static/navigationSchema.json")
	modulesFiles, err := filepath.Glob(cwd + "/static/**/**/**/*-navigation.json")
	handleErr(err)

	for _, file := range modulesFiles {
		documentLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", file))
		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		
		if err != nil {
			fmt.Println("File", file)
			panic(err.Error())
		}

		if !result.Valid() {
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
			panic(fmt.Sprintf("The %s is not valid. see errors :\n", file))
		}

		var data map[string]interface {}
		
		var arrayData [1] map[string]interface {};

		fileContent, status := ioutil.ReadFile(file)
		handleErr(status)

		ok:= json.Unmarshal(fileContent, &data)
		if ok == nil{
			//Note, only works if there are no errors unmarshalling!
			parsingSetup(data, file)
		} else {
			ok := json.Unmarshal(fileContent, &arrayData)
			if ok == nil {
				rootJSONFile := arrayData[0]
				parsingSetup(rootJSONFile, file)
			} else {
				panic(ok.Error())
			}
		}
	}
}

func parsingSetup(data map[string]interface{}, file string) {
	duplicateCounter = make(map[string]int)
	idValue, ok := data["id"]
	if ok {
		if idMap, ok := idValue.(string); ok {
			duplicateCounter[idMap] = 1
		} else {
			panic("id is not a string")
		}
	}
	parseJSONIDs(data, file)
}

func parseJSONIDs(data map[string]interface{}, file string) {
	navItems, ok := data["navItems"].([]interface{})
	if ok {
		loopOverFields(navItems, file)
	} else {
		routeItems, ok := data["routes"].([]interface{})
		if ok {
			loopOverFields(routeItems, file)
		}
	}
}

func loopOverFields(navItems []interface{}, file string) {
	for i := 0; i < len(navItems); i++ {
		navItem, ok := navItems[i].(map[string]interface{})
			if ok {
				value, ok := navItem["id"]
				if ok {
					if id, ok := value.(string); ok {
						if _, exists := duplicateCounter[id]; exists {
							panic(fmt.Sprintf("The id %s in %s is not valid because it is duplicated\n", id, file))
						} else {
							duplicateCounter[id] = 1
						}
						//FOR DEBUGGING PURPOSES
						// fmt.Println("----")
						// fmt.Println(id)
						// fmt.Println(file)
					}
				}
				parseJSONIDs(navItem, file)
			} else {
				panic(fmt.Sprintf("Invalid format. The 'navItems' field MUST be a map"))
			}
	}
}