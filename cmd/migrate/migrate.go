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

	// Wrap DB migration into a single transaction
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.AutoMigrate(&models.FavoritePage{}, &models.LastVisitedPage{}, &models.SelfReport{}, &models.UserIdentity{}, &models.ProductOfInterest{})
		if err != nil {
			return err
		}
		// Migrate all nil visited_bundles to empty object
		bundleRes = tx.Model(&models.UserIdentity{}).Where("visited_bundles IS NULL").Update("visited_bundles", []byte(`{}`))
		if bundleRes.Error != nil {
			return bundleRes.Error
		}

		return nil
	})

	if err != nil {
		logrus.Error("Unable to migrate database!")
		panic(err)
	}

	if bundleRes.RowsAffected > 0 {
		logrus.Infof("Migrated %d user identity bundles rows", bundleRes.RowsAffected)
	}
	logrus.Info("Migration complete")
}
