package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/joho/godotenv"
)

type FeatureFlag struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

const featureFlagPath = "cmd/unleash/featureflag.json"

var baseUrl string

func main() {
	godotenv.Load()
	cfg := config.Get()
	baseUrl = fmt.Sprintf("%s%s", cfg.FeatureFlagConfig.FullURL, "admin/projects/default/features")
	// Read JSON data from file
	buff, err := ioutil.ReadFile(featureFlagPath)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal JSON data into slice of structs
	var flags []FeatureFlag
	err = json.Unmarshal(buff, &flags)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range flags {
		err := createFeatureFlagEntry(v, cfg)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}
		if v.Enabled {
			err = setEnabled(v, cfg)
			if err != nil {
				fmt.Printf("Error: %v", err)
				return
			}
		}
	}

}

func createFeatureFlagEntry(f FeatureFlag, cfg *config.ChromeServiceConfig) error {
	client := &http.Client{}
	jsonData, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Send HTTP request to Unleash server
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cfg.FeatureFlagConfig.AdminToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer resp.Body.Close()

	// Process response from Unleash server
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(body))
	return nil
}

func setEnabled(f FeatureFlag, cfg *config.ChromeServiceConfig) error {
	flagUrl := fmt.Sprintf("%s/%s/environments/development/on", baseUrl, f.Name)
	client := &http.Client{}
	jsonData, err := json.Marshal(f)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// Send HTTP request to Unleash server
	req, err := http.NewRequest("POST", flagUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", cfg.FeatureFlagConfig.AdminToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("Flag %s set to enabled with code %d", f.Name, resp.StatusCode)
	return nil
}
