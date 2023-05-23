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
			for key, value:=range data {
				fmt.Println(key)
				fmt.Println(value)
				fmt.Println("---")
			}
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

	// func parseJSONIDs(data map[string]interface{}) {
	// 	hashMap := make(map[string]int)
	// 	// var paths string[]
	// 	// path, exists := data["navItems"]
	// 	// paths = append(paths, "id")
	// 	// if exists != nil {
	// 	// 	for _, str := range path {
	// 	// 		paths = append(paths, "")
	// 	// 	}
	// 	// }
	// 	hashMap[data["id"]] = 1
	// 	for i := 0; i < 2; i++ {
	// 		value, ok := data[paths[i]]
	// 		if ok != nil {
	// 			break;
	// 		}
	// 		else {
	// 			if val, status := paths[value]; status {
	// 				hashMap[value] += 1
	// 			}
	// 			else {
	// 				hashMap[value] = 1
	// 			}
	// 		}
	// 	}
	// 	string firstPath = ""
	// }
}
