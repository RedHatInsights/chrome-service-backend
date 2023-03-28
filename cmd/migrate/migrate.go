package main

import (
	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	config.Init()
	database.Init()

	var bundleRes *gorm.DB
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		logrus.Error("Unable to migrate database!")
		tx.Rollback()
		panic(err)
	}

	if err := tx.AutoMigrate(&models.FavoritePage{}, &models.LastVisitedPage{}, &models.SelfReport{}, &models.UserIdentity{}, &models.ProductOfInterest{}); err != nil {
		logrus.Error("Unable to migrate database!")
		tx.Rollback()
		panic(err)
	}

	bundleRes = tx.Model(&models.UserIdentity{}).Where("visited_bundles IS NULL").Update("visited_bundles", []byte(`{}`))
	if bundleRes.Error != nil {
		logrus.Error("Unable to migrate database!")
		tx.Rollback()
		panic(bundleRes.Error)
	}

	err := tx.Commit().Error

	if err != nil {
		logrus.Error("Unable to migrate database!")
		tx.Rollback()
		panic(err)
	}

	if bundleRes.RowsAffected > 0 {
		logrus.Infof("Migrated %d user identity bundles rows", bundleRes.RowsAffected)
	}
	logrus.Info("Migration complete")
}
