package main

import (
	"fmt"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"github.com/xeipuuv/gojsonschema"
)

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
		if ok == nil {
			// for key, value:=range data {
			// 	fmt.Println(key)
			// 	fmt.Println(value)
			// 	fmt.Println("---")
			// }
		}
		
		parseJSONIDs(data, file)

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
	hashMap := make(map[string]int)
	idValue, ok := data["id"]
	if idMap, ok := idValue.(string); ok {
		hashMap[idMap] = 1
	}
	navItems, ok := data["navItems"].([]interface{})
	if ok {
		for i := 0; i < len(navItems); i++ {
			navItem, ok := navItems[i].(map[string]interface{})
			if !ok {
				continue
			}
			value, ok := navItem["id"]
			if id, ok := value.(string); ok {
				if _, exists := hashMap[id]; exists {
					panic(fmt.Sprintf("The id %s in %s is not valid because it is duplicated\n", id, file))
				} else {
					hashMap[id] = 1
				}
				fmt.Println("----")
				fmt.Println(id)
			} else {
				continue
			}
		}
	}
}
