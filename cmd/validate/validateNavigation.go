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
		
		var data map[string]interface {}
		
		fileContent, status := ioutil.ReadFile(file)
		handleErr(status)

		ok:= json.Unmarshal(fileContent, &data)
		if ok == nil{
			//Note, only works if there are no errors unmarshalling!
			parseJSONIDs(data, file)
		}

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
	}
}

func parseJSONIDs(data map[string]interface{}, file string) {
	duplicateCounter = make(map[string]int)
	idValue, ok := data["id"]
	if idMap, ok := idValue.(string); ok {
		duplicateCounter[idMap] = 1
	}
	navItems, ok := data["navItems"].([]interface{})
	if ok {
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
						// FOR DEBUGGING PURPOSES
						// fmt.Println("----")
						// fmt.Println(id)
					}
				}
				checkForNestedItems(file, navItem)
			} 
		}
	}
}

func checkForNestedItems(file string, navItem map[string]interface{}) {
	nestedNavItems, ok := navItem["navItems"].([]interface{})
	if ok {
		for i := 0; i < len(nestedNavItems); i++ {
			nestedNavItem, ok := nestedNavItems[i].(map[string]interface{})
			if ok {
				value, ok := nestedNavItem["id"]
				if ok {
					if id, ok := value.(string); ok {
						if _, exists := duplicateCounter[id]; exists {
							panic(fmt.Sprintf("The id %s in %s is not valid because it is duplicated\n", id, file))
						} else {
							duplicateCounter[id] = 1
						}
						// FOR DEBUGGING PURPOSES
						// fmt.Println("----")
						// fmt.Println(id)
					} 
				}
				checkForNestedItems(file, nestedNavItem)
			}
		}
	} else {
		nestedRouteItems, ok := navItem["routes"].([]interface{})
		if ok {
			for i := 0; i < len(nestedRouteItems); i++ {
				nestedRouteItem, ok := nestedRouteItems[i].(map[string]interface{})
				if ok {
					value, ok := nestedRouteItem["id"]
					if ok {
						if idRoute, ok := value.(string); ok {
							if _, exists := duplicateCounter[idRoute]; exists {
								panic(fmt.Sprintf("The id %s in %s is not valid because it is duplicated\n", idRoute, file))
							} else {
								duplicateCounter[idRoute] = 1
							}
							// FOR DEBUGGING PURPOSES
							// fmt.Println("----")
							// fmt.Println(idRoute)
						} 
					}
					checkForNestedItems(file, nestedRouteItem)
				}
			}
		} 
	}
}