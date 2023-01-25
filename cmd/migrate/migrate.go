package main

import (
	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load()
	config.Init()
	database.Init()
	err := database.DB.AutoMigrate(&models.FavoritePage{}, &models.LastVisitedPage{}, &models.SelfReport{}, &models.UserIdentity{}, &models.ProductOfInterest{})
	if err != nil {
		panic(err)
	}
	logrus.Info("Migration complete")
}
