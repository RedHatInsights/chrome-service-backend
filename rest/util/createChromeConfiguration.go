package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const (
	fedModulesPath       = "static/fed-modules-generated.json"
	staticFedModulesPath = "static/stable/%s/modules/fed-modules.json"
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
	// searchVar := os.Getenv("FEO_SEARCH_INDEX")
	// serviceTilesVar := os.Getenv("FEO_SERVICE_TILES")
	// widgetRegistryVar := os.Getenv("FEO_WIDGET_REGISTRY")

	fedModules, err := parseFedModules(fedModulesVar, env)
	if err != nil {
		panic(fmt.Sprintf("Error parsing FEO_FED_MODULES: %v", err))
	}

	fmt.Println("Writing fed-modules.json")
	err = writeConfigFile(fedModules, fedModulesPath)
	if err != nil {
		panic(err)
	}
}
