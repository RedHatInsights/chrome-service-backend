package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type environmentsLinkStorage struct {
	Itless map[string]interface{} `json:"itless,omitempty"`
	Stage  map[string]interface{} `json:"stage,omitempty"`
	Prod   map[string]interface{} `json:"prod,omitempty"`
}

type releaseLinksStorage struct {
	Beta   environmentsLinkStorage `json:"beta,omitempty"`
	Stable environmentsLinkStorage `json:"stable,omitempty"`
}

var (
	releases        []string            = []string{"beta", "stable"}
	environments    []string            = []string{"itless", "prod", "stage"}
	navLinksStorage releaseLinksStorage = releaseLinksStorage{
		Beta: environmentsLinkStorage{
			Itless: map[string]interface{}{},
			Stage:  map[string]interface{}{},
			Prod:   map[string]interface{}{},
		},
		Stable: environmentsLinkStorage{
			Itless: map[string]interface{}{},
			Stage:  map[string]interface{}{},
			Prod:   map[string]interface{}{},
		},
	}
)

type templateBase struct {
	Id          string        `json:"id,omitempty"`
	Icon        string        `json:"icon,omitempty"`
	Title       string        `json:"title,omitempty"`
	Description string        `json:"description,omitempty"`
	Links       []interface{} `json:"links,omitempty"`
}

type serviceItem map[string]interface{}

func parseServiceItemLink(link interface{}, linkStorage *map[string]interface{}) map[string]interface{} {
	l, ok := link.(map[string]interface{})
	if !ok {
		// can be a string
		linkId, ok := link.(string)
		// TODO: Extract link from navigation files
		if !ok {
			panic("Unable to parse link to string map")
		}

		linkStruct, ok := (*linkStorage)[linkId].(map[string]interface{})
		if !ok {
			panic(fmt.Errorf("unable to find link with id %s", linkId))
		}
		return linkStruct
	}

	if l["isGroup"] == true {
		rawGroupLinks, ok := l["links"].([]interface{})
		if !ok {
			panic("Unable to parse group links")
		}
		var groupLinks []map[string]interface{}
		for _, groupLink := range rawGroupLinks {
			gl, ok := groupLink.(map[string]interface{})
			if !ok {
				// probably a string
				_, ok := groupLink.(string)
				if !ok {
					panic("Unable to parse group link to string map")
				}
				groupLinks = append(groupLinks, parseServiceItemLink(groupLink, linkStorage))
			} else {
				groupLinks = append(groupLinks, gl)
			}
		}
		l["links"] = groupLinks
	}

	return l

}

func parseServiceItem(tb templateBase, linkStorage *map[string]interface{}) serviceItem {
	serviceItem := make(serviceItem)
	serviceItem["id"] = tb.Id
	serviceItem["icon"] = tb.Icon
	serviceItem["title"] = tb.Title
	serviceItem["description"] = tb.Description
	var serviceItemLinks []interface{}
	for _, l := range tb.Links {
		link := parseServiceItemLink(l, linkStorage)
		serviceItemLinks = append(serviceItemLinks, link)
	}
	serviceItem["links"] = serviceItemLinks
	return serviceItem
}

func parseEnvironment(assetPath string, linkStorage *map[string]interface{}) []serviceItem {
	var output []serviceItem
	var fileData []templateBase
	file, err := os.ReadFile(assetPath)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, &fileData)
	if err != nil {
		panic(err)
	}

	for _, f := range fileData {
		output = append(output, parseServiceItem(f, linkStorage))
	}

	return output
}

type navigationTemplate struct {
	NavItems []map[string]interface{} `json:"navItems,omitempty"`
}

func getLinksStorage(release string, environment string) *map[string]interface{} {
	if release == "stable" {
		switch environment {
		case "itless":
			return &navLinksStorage.Stable.Itless
		case "prod":
			return &navLinksStorage.Stable.Prod
		case "stage":
			return &navLinksStorage.Stable.Stage
		default:
			panic(fmt.Errorf("unknown environment %s", environment))
		}
	} else if release == "beta" {
		switch environment {
		case "itless":
			return &navLinksStorage.Beta.Itless
		case "prod":
			return &navLinksStorage.Beta.Prod
		case "stage":
			return &navLinksStorage.Beta.Stage
		default:
			panic(fmt.Errorf("unknown environment %s", environment))
		}
	} else {
		panic(fmt.Errorf("unknown release %s", release))
	}
}

func parseNestedLinks(nestedLink []interface{}) []map[string]interface{} {
	var flatLinks []map[string]interface{}
	for _, link := range nestedLink {
		l, ok := link.(map[string]interface{})
		if !ok {
			panic("unable to parse nested link")
		}
		flatLinks = append(flatLinks, l)
	}
	return flatLinks

}

func findFirstValidLeaf(navItems []map[string]interface{}) (map[string]interface{}, bool) {
	var leaf map[string]interface{}
	found := false
	for _, item := range navItems {
		stepNavItems, niOk := item["navItems"].([]interface{})
		_, eOk := item["expandable"].(bool)
		stepRoutes, rOk := item["routes"].([]interface{})
		var routes []map[string]interface{}
		var nestedNavItems []map[string]interface{}
		if rOk {
			routes = parseNestedLinks(stepRoutes)
		}

		if niOk {
			nestedNavItems = parseNestedLinks(stepNavItems)
		}

		if rOk && eOk && len(routes) > 0 {
			leaf, found = findFirstValidLeaf(routes)
		} else if niOk && len(nestedNavItems) > 0 {
			leaf, found = findFirstValidLeaf(nestedNavItems)
		} else {
			leaf = item
			found = true
		}

		if found {
			break
		}
	}
	return leaf, found
}

func parseNavigationLinks(navItems []map[string]interface{}) []map[string]interface{} {
	var flatItems []map[string]interface{}
	for _, item := range navItems {

		idOk := true
		stepNavItems, niOk := item["navItems"].([]interface{})
		_, eOk := item["expandable"].(bool)
		stepRoutes, rOk := item["routes"].([]interface{})
		var routes []map[string]interface{}
		var nestedNavItems []map[string]interface{}
		if rOk {
			routes = parseNestedLinks(stepRoutes)
		}

		if niOk {
			nestedNavItems = parseNestedLinks(stepNavItems)
		}

		if item["id"] != nil {
			_, idOk = item["id"].(string)
			if niOk && idOk {
				leafItem, found := findFirstValidLeaf(nestedNavItems)
				leaf := make(map[string]interface{})
				if !found {
					panic(fmt.Errorf("unable to find leaf for %v", item))
				}
				for k, v := range leafItem {
					leaf[k] = v
				}
				// override key attributes
				leaf["id"] = item["id"]
				leaf["title"] = item["title"]
				leaf["description"] = item["description"]
				flatItems = append(flatItems, leaf)
			}

			if eOk && idOk {
				leafItem, found := findFirstValidLeaf(routes)
				leaf := make(map[string]interface{})
				if !found {
					panic(fmt.Errorf("unable to find leaf for %v", item))
				}
				for k, v := range leafItem {
					leaf[k] = v
				}
				// override key attributes
				leaf["id"] = item["id"]
				leaf["title"] = item["title"]
				leaf["description"] = item["description"]
				flatItems = append(flatItems, leaf)
			}
		}
		if niOk && len(nestedNavItems) > 0 {
			// a group branch
			flatItems = append(flatItems, parseNavigationLinks(nestedNavItems)...)
		} else if rOk && eOk && len(routes) > 0 {
			// expandable section
			flatItems = append(flatItems, parseNavigationLinks(routes)...)
		} else if idOk && item["id"] != nil {
			// normal link
			flatItems = append(flatItems, item)
		} else if item["id"] != nil {
			panic(fmt.Errorf("unable to parse navigation item %v", item))
		}
	}
	return flatItems
}

func parseEnvironmentLinks(navFilePath string, linkStorage *map[string]interface{}) {
	var navStruct navigationTemplate
	navFile, err := os.ReadFile(navFilePath)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to read navigation file %s", navFilePath))
		panic(err)
	}

	err = json.Unmarshal(navFile, &navStruct)
	if err != nil {
		fmt.Println(navFilePath)
		panic(err)
	}

	fileSegments := strings.Split(filepath.Base(navFilePath), "/")
	bundleId := strings.Split(fileSegments[len(fileSegments)-1], "-navigation.json")[0]

	flatLinks := parseNavigationLinks(navStruct.NavItems)
	for _, link := range flatLinks {
		serviceLinkId := fmt.Sprintf("%s.%s", bundleId, link["id"].(string))
		if (*linkStorage)[serviceLinkId] != nil {
			panic(fmt.Errorf("duplicate link id %s", serviceLinkId))
		}
		(*linkStorage)[serviceLinkId] = link
	}

}

func parseAllEnvironments() {
	for _, release := range releases {
		for _, environment := range environments {
			fmt.Printf("Parsing %s %s\n", release, environment)
			servicesTemplatePath := fmt.Sprintf("static/%s/%s/services/services.json", release, environment)
			navFilesPaths, err := filepath.Glob(fmt.Sprintf("static/%s/%s/navigation/*-navigation.json", release, environment))
			linkStorage := getLinksStorage(release, environment)
			if err != nil || len(navFilesPaths) == 0 {
				panic("unable to find navigation files")
			}
			for _, nfp := range navFilesPaths {
				if !strings.Contains(nfp, "landing") {
					parseEnvironmentLinks(nfp, linkStorage)
				}
			}
			storageFileName := fmt.Sprintf("static/%s/%s/services/links-storage.json", release, environment)
			storageFileContent, err := json.MarshalIndent(linkStorage, "", "  ")
			if err != nil {
				panic(err)
			}
			os.WriteFile(storageFileName, storageFileContent, 0644)

			output := parseEnvironment(servicesTemplatePath, linkStorage)
			fileName := fmt.Sprintf("static/%s/%s/services/services-generated.json", release, environment)
			fileContent, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				panic(err)
			}
			os.WriteFile(fileName, fileContent, 0644)
		}
	}
}

func main() {
	parseAllEnvironments()
}
