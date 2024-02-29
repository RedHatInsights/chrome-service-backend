package main

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()
	database.Init()

	var bundleRes *gorm.DB
	var visitedRes *gorm.DB
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

	// fk_user_identities_last_visited_pages
	if tx.Migrator().HasConstraint(&models.UserIdentity{}, "fk_user_identities_last_visited_pages") {
		if err := tx.Migrator().DropConstraint(&models.UserIdentity{}, "fk_user_identities_last_visited_pages"); err != nil {
			logrus.Error("Unable to migrate database!")
			tx.Rollback()
			panic(err)
		}
	}

	// temporary - removes unused typo column in dashboard template tables
	if tx.Migrator().HasColumn(&models.DashboardTemplate{}, "sx") {
		if err := tx.Migrator().DropColumn(&models.DashboardTemplate{}, "sx"); err != nil {
			logrus.Error("Unable to migrate database!")
			tx.Rollback()
			panic(err)
		}
	}

	if err := tx.AutoMigrate(&models.FavoritePage{}, &models.UserIdentity{}, &models.SelfReport{}, &models.ProductOfInterest{}, &models.DashboardTemplate{}); err != nil {
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
	visitedRes = tx.Model(&models.UserIdentity{}).Where("last_visited_pages IS NULL").Update("last_visited_pages", []byte(`[]`))
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

	if visitedRes.RowsAffected > 0 {
		logrus.Infof("Migrated %d user identity visited rows", visitedRes.RowsAffected)
	}
	logrus.Info("Migration complete")
}
