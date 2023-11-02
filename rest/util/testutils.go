package util

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
)

const ProjectName = "chrome-service-backend"

func LoadEnv() error {
	projectName := regexp.MustCompile(`^(.*` + ProjectName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))

	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Println("Error loading custom .env file. Falling back to standard")
		return err
	}
	return nil
}
