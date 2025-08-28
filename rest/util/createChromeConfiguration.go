package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const (
	fedModulesPath        = "static/fed-modules-generated.json"
	staticFedModulesPath  = "static/stable/%s/modules/fed-modules.json"
	searchIndexPath       = "static/search-index-generated.json"
	staticNavigationFiles = "static/stable/%s/navigation"
	bundlesPath           = "static/bundles-generated.json"
	staticServicesPath    = "static/stable/%s/services/services-generated.json"
	serviceTilesPath      = "static/service-tiles-generated.json"
	apiSpecPath           = "static/api-spec.json"
)

func getLegacyConfigFile(path string, env string) ([]byte, error) {
	// read the file
	file, err := os.ReadFile(fmt.Sprintf(path, env))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func parseFedModules(fedModulesConfig string, env string) ([]byte, error) {

	fm := make(map[string]interface{})
	lfm := make(map[string]interface{})
	if fedModulesConfig == "" {
		logrus.Warn("FEO_FED_MODULES is not set, using empty configuration")
	} else {
		err := json.Unmarshal([]byte(fedModulesConfig), &fm)
		// do the thing
		if err != nil {
			return nil, err
		}
	}

	// read legacy configFile
	legacyFedModulesFile, err := getLegacyConfigFile(staticFedModulesPath, env)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(legacyFedModulesFile, &lfm)
	if err != nil {
		return nil, err
	}

	for key, value := range fm {
		// We will have to wrap this assignment to some kind of if statement to make sure we don't overwrite the values if the frontend resource is not ready to be used
		// merge legacy and generated values
		lfm[key] = value
	}

	// parse back to string so it can be written to a file
	res, err := json.MarshalIndent(lfm, "", "  ")
	return res, err
}

func parseSearchIndex(searchIndexConfig string, env string) ([]byte, error) {
	si := []interface{}{}

	if searchIndexConfig == "" {
		logrus.Warn("FEO_SEARCH_INDEX is not set, using empty configuration")
	} else {
		err := json.Unmarshal([]byte(searchIndexConfig), &si)
		if err != nil {
			return nil, err
		}
	}

	res, err := json.MarshalIndent(si, "", "  ")
	return res, err
}

func writeConfigFile(config []byte, path string) error {
	cwd, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	file := fmt.Sprintf("%s/%s", cwd, path)
	logrus.Infof("Writing configuration to %s", file)
	err = os.WriteFile(file, config, 0644)
	return err
}

func getLegacyNavFiles(env string) ([]string, error) {
	navigationFiles, err := filepath.Glob(fmt.Sprintf(staticNavigationFiles, env) + "/*-navigation.json")
	return navigationFiles, err
}

func createLegacyBundles(navigationFiles []string) ([]interface{}, error) {
	bundles := []interface{}{}
	for _, fileName := range navigationFiles {
		if strings.Contains(fileName, "landing-navigation.json") {
			continue
		}
		file, err := os.ReadFile(fileName)
		if err != nil {
			return bundles, err
		}
		var bundleEntry interface{}
		err = json.Unmarshal(file, &bundleEntry)
		if err != nil {
			return bundles, err
		}
		bundles = append(bundles, bundleEntry)
	}

	return bundles, nil
}

func replaceNavItem(navItem map[string]interface{}, availableReplacements map[string]map[string]interface{}) (map[string]interface{}, error) {
	if navItem["feoReplacement"] == nil {
		if navItem["routes"] != nil {
			serializedRoutes, err := json.Marshal(navItem["routes"])
			if err != nil {
				return nil, err
			}
			nestedRoutes := []map[string]interface{}{}
			err = json.Unmarshal(serializedRoutes, &nestedRoutes)
			if err != nil {
				return nil, err
			}
			replacementRoutes := []map[string]interface{}{}
			for _, route := range nestedRoutes {
				replacementRoute, err := replaceNavItem(route, availableReplacements)
				if err != nil {
					return nil, err
				}
				replacementRoutes = append(replacementRoutes, replacementRoute)
			}
			navItem["routes"] = replacementRoutes
		}

		if navItem["navItems"] != nil {
			serializedNavItems, err := json.Marshal(navItem["navItems"])
			if err != nil {
				return nil, err
			}
			nestedNavItems := []map[string]interface{}{}
			err = json.Unmarshal(serializedNavItems, &nestedNavItems)
			if err != nil {
				return nil, err
			}
			replacementNavItems := []map[string]interface{}{}
			for _, navItem := range nestedNavItems {
				replacementNavItem, err := replaceNavItem(navItem, availableReplacements)
				if err != nil {
					return nil, err
				}
				replacementNavItems = append(replacementNavItems, replacementNavItem)
			}
			navItem["navItems"] = replacementNavItems
		}
		return navItem, nil
	}

	replacementId, ok := navItem["feoReplacement"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid feoReplacement id type")
	}
	replacement := availableReplacements[replacementId]
	if replacement == nil {
		logrus.Warnln("Replacement not found for: ", replacementId)
		return navItem, nil
	}

	return replacement, nil
}

func createAvailableReplacements(generatedNavItems []map[string]interface{}) (map[string]map[string]interface{}, error) {
	availableReplacements := map[string]map[string]interface{}{}
	for _, navItem := range generatedNavItems {
		replacementId, ok := navItem["id"].(string)
		if !ok {
			logrus.Warnln("Invalid navItem id type")
			continue
		}
		availableReplacements[replacementId] = navItem
		if navItem["navItems"] != nil {
			serializedNavItems, err := json.Marshal(navItem["navItems"])
			if err != nil {
				return nil, err
			}
			nestedNavItems := []map[string]interface{}{}
			err = json.Unmarshal(serializedNavItems, &nestedNavItems)
			if err != nil {
				return nil, err
			}
			for _, navItem := range nestedNavItems {
				nestedNavItem, err := replaceNavItem(navItem, availableReplacements)
				if err != nil {
					return nil, err
				}
				nestedNavItems = append(nestedNavItems, nestedNavItem)
			}
			nestedReplacements, err := createAvailableReplacements(nestedNavItems)
			if err != nil {
				return nil, err
			}
			for key, value := range nestedReplacements {
				availableReplacements[key] = value
			}
		}

		if navItem["routes"] != nil {
			serializedRoutes, err := json.Marshal(navItem["routes"])
			if err != nil {
				return nil, err
			}
			nestedRoutes := []map[string]interface{}{}
			err = json.Unmarshal(serializedRoutes, &nestedRoutes)
			if err != nil {
				return nil, err
			}
			for _, route := range nestedRoutes {
				nestedRoute, err := replaceNavItem(route, availableReplacements)
				if err != nil {
					return nil, err
				}
				nestedRoutes = append(nestedRoutes, nestedRoute)
			}
			nestedReplacements, err := createAvailableReplacements(nestedRoutes)
			if err != nil {
				return nil, err
			}
			for key, value := range nestedReplacements {
				availableReplacements[key] = value
			}
		}
	}
	return availableReplacements, nil
}

func replaceBundleItems(bundle map[string]interface{}, generatedBundle map[string]interface{}) (map[string]interface{}, error) {
	replacedBundle := map[string]interface{}{}
	serializedGeneratedNavItems, err := json.Marshal(generatedBundle["navItems"])
	if err != nil {
		return nil, err
	}
	generatedNavItems := []map[string]interface{}{}
	json.Unmarshal(serializedGeneratedNavItems, &generatedNavItems)
	availableReplacements, err := createAvailableReplacements(generatedNavItems)
	if err != nil {
		return nil, err
	}

	serializedOriginalNavItems, err := json.Marshal(bundle["navItems"])
	if err != nil {
		return nil, err
	}
	convertedNavItems := []map[string]interface{}{}
	json.Unmarshal(serializedOriginalNavItems, &convertedNavItems)

	replacementNavItems := []map[string]interface{}{}
	for _, navItem := range convertedNavItems {
		replacementNavItem, err := replaceNavItem(navItem, availableReplacements)
		if err != nil {
			return nil, err
		}
		replacementNavItems = append(replacementNavItems, replacementNavItem)
	}

	replacedBundle["navItems"] = replacementNavItems
	replacedBundle["id"] = bundle["id"]
	replacedBundle["title"] = bundle["title"]

	return replacedBundle, nil
}

func parseBundles(bundlesVar string, bundlesOnboardedIdsVar string, env string) ([]byte, error) {
	navigationFiles, err := getLegacyNavFiles(env)
	generatedBundles := []interface{}{}
	onboardedBundleIds := []interface{}{}
	if bundlesVar != "" {
		err := json.Unmarshal([]byte(bundlesVar), &generatedBundles)
		if err != nil {
			return nil, err
		}
	}
	if bundlesOnboardedIdsVar != "" {
		err := json.Unmarshal([]byte(bundlesOnboardedIdsVar), &onboardedBundleIds)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	bundles, err := createLegacyBundles(navigationFiles)
	if err != nil {
		return nil, err
	}
	parsedBundles := []map[string]interface{}{}
	for _, bundle := range bundles {
		mapBundle, ok := bundle.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid bundle type")
		}
		skipMerge := false
		for _, onboardedBundleId := range onboardedBundleIds {
			bundleId, ok := onboardedBundleId.(string)
			if !ok {
				return nil, fmt.Errorf("invalid onboarded bundle ID type")
			}
			if bundleId == mapBundle["id"] {
				skipMerge = true
				break
			}
		}
		for _, generatedBundle := range generatedBundles {
			generatedAlternative, ok := generatedBundle.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid generated bundle type")
			}
			if generatedAlternative["id"] == mapBundle["id"] {
				if skipMerge {
					mapBundle = generatedAlternative
					break
				}
				mapBundle, err = replaceBundleItems(mapBundle, generatedAlternative)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		parsedBundles = append(parsedBundles, mapBundle)
	}
	bundleData, err := json.MarshalIndent(parsedBundles, "", "  ")
	return bundleData, err
}

type sectionGroup struct {
	ID      string                   `json:"id"`
	Title   string                   `json:"title"`
	IsGroup bool                     `json:"isGroup"`
	Links   []map[string]interface{} `json:"links"`
}

type serviceSection struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Icon        string         `json:"icon"`
	Title       string         `json:"title"`
	Links       []sectionGroup `json:"links"`
}

type feoServiceGroup struct {
	ID    string                   `json:"id"`
	Title string                   `json:"title"`
	Tiles []map[string]interface{} `json:"tiles"`
}
type feoServiceSection struct {
	Icon        string            `json:"icon"`
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Groups      []feoServiceGroup `json:"groups"`
}

func parseServiceTiles(serviceTilesVar string, env string) ([]byte, error) {
	legacyServicesFile, err := getLegacyConfigFile(staticServicesPath, env)
	if err != nil {
		logrus.Errorln("Error reading legacy services file")
		return nil, err
	}
	lsf := []serviceSection{}
	nsf := []feoServiceSection{}
	// for easier replacement
	nsfAccessMap := map[string]feoServiceSection{}
	err = json.Unmarshal(legacyServicesFile, &lsf)
	if err != nil {
		logrus.Errorln("Error parsing legacy services file")
		return nil, err
	}

	if serviceTilesVar != "" {
		err := json.Unmarshal([]byte(serviceTilesVar), &nsf)
		if err != nil {
			logrus.Errorln("Error parsing service tiles env variable")
			return nil, err
		}
	} else {
		logrus.Infoln("No service tiles env variable found")
	}
	for _, service := range nsf {
		sectionId := service.ID
		nsfAccessMap[sectionId] = service
	}

	newServices := []serviceSection{}
	// replace legacy services with new services, append new services if not found in legacy
	for _, service := range lsf {
		sectionId := service.ID
		generatedSection := nsfAccessMap[sectionId]
		if generatedSection.ID == "" {
			// this service was not generated, skip
			// newServices = append(newServices, service)
			continue
		}
		service.Description = generatedSection.Description
		service.Icon = generatedSection.Icon
		service.Title = generatedSection.Title
		legacySectionGroups := service.Links
		generatedSectionLinksAccessMap := map[string]feoServiceGroup{}
		for _, group := range generatedSection.Groups {
			generatedSectionLinksAccessMap[group.ID] = group
		}

		newSectionGroups := []sectionGroup{}
		for _, group := range legacySectionGroups {
			groupId := group.ID
			generatedGroup := generatedSectionLinksAccessMap[groupId]
			if generatedGroup.ID == "" {
				// this group was not generated, skip
				newSectionGroups = append(newSectionGroups, group)
				continue
			}
			group.Title = generatedGroup.Title

			newGroupLinks := []map[string]interface{}{}
			generatedGroupLinksAccessMap := map[string]map[string]interface{}{}
			for _, link := range generatedGroup.Tiles {
				linkId, ok := link["id"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid generated link id type, section %s, group %s", sectionId, groupId)
				}
				generatedGroupLinksAccessMap[linkId] = link
			}

			for _, link := range group.Links {
				linkId, ok := link["id"].(string)
				if !ok {
					logrus.Warningf("invalid legacy link id type, section %s, group %s", sectionId, groupId)
					newGroupLinks = append(newGroupLinks, link)
					continue
				}
				if generatedGroupLinksAccessMap[linkId] == nil {
					newGroupLinks = append(newGroupLinks, link)
				} else {
					newGroupLinks = append(newGroupLinks, generatedGroupLinksAccessMap[linkId])
					delete(generatedGroupLinksAccessMap, linkId)
				}
			}
			for _, link := range generatedGroupLinksAccessMap {
				// add links not defined in legacy
				newGroupLinks = append(newGroupLinks, link)
			}
			group.Links = newGroupLinks
			delete(generatedSectionLinksAccessMap, groupId)
			newSectionGroups = append(newSectionGroups, group)
		}

		service.Links = newSectionGroups
		newServices = append(newServices, service)
		delete(nsfAccessMap, sectionId)
	}

	// add sections not defined in static files
	for _, section := range nsfAccessMap {
		newGroups := []sectionGroup{}
		for _, group := range section.Groups {
			newGroup := sectionGroup{
				ID:      group.ID,
				Title:   group.Title,
				IsGroup: true,
				Links:   group.Tiles,
			}
			newGroups = append(newGroups, newGroup)
		}
		newSection := serviceSection{
			ID:          section.ID,
			Description: section.Description,
			Icon:        section.Icon,
			Title:       section.Title,
			Links:       newGroups,
		}
		newServices = append(newServices, newSection)
	}

	servicesByte, err := json.MarshalIndent(newServices, "", "  ")
	if err != nil {
		logrus.Errorln("Error marshalling new services")
		return nil, err
	}

	return servicesByte, nil
}

func parseApiSpec(apiSpecVar string, env string) ([]byte, error) {
	var apiSpec interface{}

	if apiSpecVar == "" {
		logrus.Warn("FEO_API_SPEC is not set, using empty configuration")
		apiSpec = map[string]interface{}{}
	} else {
		err := json.Unmarshal([]byte(apiSpecVar), &apiSpec)
		if err != nil {
			return nil, err
		}
	}

	res, err := json.MarshalIndent(apiSpec, "", "  ")
	return res, err
}

func CreateChromeConfiguration() {
	// These parsing methods are temporary due to a longer migration window offered to UI tenants
	// Once migrated, most of the parsing will be removed and the config files will be simply forwarded to the UI tenants
	err := LoadEnv()
	if err != nil {
		godotenv.Load()
	}

	// environment type
	// Can be one if prod, stage, itless, anything else will cause exception
	env := os.Getenv("FRONTEND_ENVIRONMENT")
	if env == "" {
		logrus.Warn("FRONTEND_ENVIRONMENT is not set, using 'stage'")
		env = "stage"
	} else if env != "prod" && env != "stage" && env != "itless" {
		panic(fmt.Sprintf("Invalid FRONTEND_ENVIRONMENT value: %s", env))
	}

	fedModulesVar := os.Getenv("FEO_FED_MODULES")
	searchVar := os.Getenv("FEO_SEARCH_INDEX")
	serviceTilesVar := os.Getenv("FEO_SERVICE_TILES")
	// widgetRegistryVar := os.Getenv("FEO_WIDGET_REGISTRY")
	bundlesVar := os.Getenv("FEO_BUNDLES")
	bundlesOnboardedIdsVar := os.Getenv("FEO_BUNDLES_ONBOARDED_IDS")
	apiSpecVar := os.Getenv("FEO_API_SPEC")

	fedModules, err := parseFedModules(fedModulesVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_FED_MODULES: %v", err))
	}

	fmt.Println("Writing fed-modules.json")
	err = writeConfigFile(fedModules, fedModulesPath)
	if err != nil {
		panic(err)
	}

	searchIndex, err := parseSearchIndex(searchVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_SEARCH_INDEX: %v", err))
	}

	err = writeConfigFile(searchIndex, searchIndexPath)
	if err != nil {
		panic(err)
	}

	bundles, err := parseBundles(bundlesVar, bundlesOnboardedIdsVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_BUNDLES: %v", err))
	}

	err = writeConfigFile(bundles, bundlesPath)
	if err != nil {
		panic(err)
	}

	services, err := parseServiceTiles(serviceTilesVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_SERVICE_TILES: %v", err))
	}

	err = writeConfigFile(services, serviceTilesPath)
	if err != nil {
		panic(err)
	}

	apiSpec, err := parseApiSpec(apiSpecVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_API_SPEC: %v", err))
	}

	err = writeConfigFile(apiSpec, apiSpecPath)
	if err != nil {
		panic(err)
	}
}
