package main

import (
	"strings"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func getValidGridItems(items []models.GridItem) []models.GridItem {
	newGridItems := []models.GridItem{}
	for _, item := range items {
		widgetTypes := strings.Split(item.ID, "#")
		if len(widgetTypes) > 0 {
			widgetType := widgetTypes[0]
			err := models.AvailableWidgets(widgetType).IsValid()
			if err == nil {
				newGridItems = append(newGridItems, item)
			} else {
				logrus.Infof("Removing invalid widget type: %s", widgetType)
			}
		}
	}

	return newGridItems
}

func migrateDashboardWidgets(tx *gorm.DB) *gorm.DB {
	var templates []models.DashboardTemplate
	res := tx.Find(&templates)
	if res.Error != nil {
		return res
	}

	for _, t := range templates {
		t.TemplateConfig.Xl = datatypes.NewJSONType(getValidGridItems(t.TemplateConfig.Xl.Data()))
		t.TemplateConfig.Lg = datatypes.NewJSONType(getValidGridItems(t.TemplateConfig.Lg.Data()))
		t.TemplateConfig.Md = datatypes.NewJSONType(getValidGridItems(t.TemplateConfig.Md.Data()))
		t.TemplateConfig.Sm = datatypes.NewJSONType(getValidGridItems(t.TemplateConfig.Sm.Data()))
		res = tx.Save(&t)

		if res.Error != nil {
			return res
		}
	}

	return res
}

func main() {
	godotenv.Load()
	database.Init()

	var bundleRes *gorm.DB
	var visitedRes *gorm.DB
	var dashboardRes *gorm.DB
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

	dashboardRes = migrateDashboardWidgets(tx)
	if dashboardRes.Error != nil {
		logrus.Error("Unable to migrate database!")
		tx.Rollback()
		panic(dashboardRes.Error)
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
