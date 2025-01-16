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
			for _, route := range nestedRoutes {
				nestedRoute, err := replaceNavItem(route, availableReplacements)
				if err != nil {
					return nil, err
				}
				nestedRoutes = append(nestedRoutes, nestedRoute)
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
			for _, navItem := range nestedNavItems {
				nestedNavItem, err := replaceNavItem(navItem, availableReplacements)
				if err != nil {
					return nil, err
				}
				nestedNavItems = append(nestedNavItems, nestedNavItem)
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

func CreateChromeConfiguration() {
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
	// serviceTilesVar := os.Getenv("FEO_SERVICE_TILES")
	// widgetRegistryVar := os.Getenv("FEO_WIDGET_REGISTRY")
	bundlesVar := os.Getenv("FEO_BUNDLES")
	bundlesOnboardedIdsVar := os.Getenv("FEO_BUNDLES_ONBOARDED_IDS")

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
}
