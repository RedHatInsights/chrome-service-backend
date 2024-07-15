package util

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
)

const ProjectName = "chrome-service-backend"

func LoadEnv() error {
	projectName := regexp.MustCompile(fmt.Sprintf(`^(.*(%s|source))`, ProjectName))
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	var err error
	if string(rootPath) == "" {
		// try fallback to default
		err = godotenv.Load()
	} else {
		err = godotenv.Load(string(rootPath) + `/.env`)
	}

	if err != nil {
		log.Println("Error loading custom .env file. Falling back to standard", err)
		return err
	}
	return nil
}
