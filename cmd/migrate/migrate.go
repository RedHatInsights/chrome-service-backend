package main

import (
	"fmt"
	"strings"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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

	logrus.Infoln("transform grid items into gorm json type")
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

	templates = []models.DashboardTemplate{}

	fmt.Println("Removing all non default templates")
	// remove all non default templates
	res = tx.Unscoped().Debug().Clauses(clause.Locking{
		Strength: "SHARE",
		Options:  "NOWAIT",
	}).Where(`"default" = ?`, false).Delete(&[]models.DashboardTemplate{})
	if res.Error != nil {
		return res
	}

	return res
}

func main() {
	godotenv.Load()
	database.Init()

	var bundleRes *gorm.DB
	var visitedRes *gorm.DB
	var dashboardRes *gorm.DB
	tx := database.DB.Begin().Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Info),
	})
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Unable to migrate database!")
			fmt.Println(r)
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		fmt.Println("Unable to migrate database!", err.Error())
		tx.Rollback()
		panic(err)
	}

	fmt.Println("Migrating last visited pages to user identity table")
	// fk_user_identities_last_visited_pages
	if tx.Migrator().HasConstraint(&models.UserIdentity{}, "fk_user_identities_last_visited_pages") {
		if err := tx.Migrator().DropConstraint(&models.UserIdentity{}, "fk_user_identities_last_visited_pages"); err != nil {
			fmt.Println("Unable to migrate database!", err.Error())
			tx.Rollback()
			panic(err)
		}
	}

	fmt.Println("Removing unfavorited pages")
	//removes unfavorited pages from all users in favorite pages tables
	if tx.Migrator().HasTable(&models.FavoritePage{}) {
		err := tx.Where(`"favorite" = ?`, false).Unscoped().Delete(&models.FavoritePage{}).Error
		if err != nil {
			fmt.Println("Unable to migrate database!", err)
			tx.Rollback()
			panic(err)
		}
	}

	fmt.Println("Removing dashboard templates sx variant")
	// temporary - removes unused typo column in dashboard template tables
	if tx.Migrator().HasColumn(&models.DashboardTemplate{}, "sx") {
		if err := tx.Migrator().DropColumn(&models.DashboardTemplate{}, "sx"); err != nil {
			fmt.Println("Unable to migrate database!", err.Error())
			tx.Rollback()
			panic(err)
		}
	}

	fmt.Println("Remove duplicate identities")
	//Deletes Duplicate users from users table
	if tx.Migrator().HasTable(&models.UserIdentity{}) {
		var duplicates []models.UserIdentity
		err := tx.Model(&models.UserIdentity{}).Select("account_id").Group("account_id").Having("COUNT(*) > 1").Limit(500).Find(&duplicates)
		if err.Error != nil {
			fmt.Println("Unable to migrate database!", err.Error.Error())
			tx.Rollback()
			panic(err.Error)
		}

		for _, dup := range duplicates {

			var usersToDelete []models.UserIdentity
			tx.Where("account_id = ?", dup.AccountId).Order("updated_at DESC").Find(&usersToDelete)
			for i, user := range usersToDelete {
				if i > 0 { // Skip the first entry, delete all others
					if err := tx.Unscoped().Where(&models.FavoritePage{
						UserIdentityID: user.ID,
					}).Delete(&models.FavoritePage{}).Error; err != nil {
						tx.Rollback()
						fmt.Println("Unable to delete user favorite pages associations!", err.Error())
						panic(err)
					}

					if err := tx.Unscoped().Where(&models.DashboardTemplate{
						UserIdentityID: user.ID,
					}).Delete(&models.DashboardTemplate{}).Error; err != nil {
						tx.Rollback()
						fmt.Println("Unable to delete user dashboard template associations!", err.Error())
						panic(err)
					}
					if err := tx.Unscoped().Delete(&user).Error; err != nil {
						tx.Rollback()
						fmt.Println("Unable to delete duplicate users!", err.Error())
						panic(err)
					}
				}
			}
		}
	}

	fmt.Println("Auto migrate relations")
	if err := tx.AutoMigrate(&models.FavoritePage{}, &models.UserIdentity{}, &models.SelfReport{}, &models.ProductOfInterest{}, &models.DashboardTemplate{}); err != nil {
		fmt.Println("Unable to migrate database!", err)
		tx.Rollback()
		panic(err)
	}

	fmt.Println("Remove last visited pages table")
	// Drop old tables
	if tx.Migrator().HasTable("last_visited_pages") {
		if err := tx.Migrator().DropTable("last_visited_pages"); err != nil {
			fmt.Println("Unable to migrate database!", err.Error())
			tx.Rollback()
			panic(err)
		}
	}

	fmt.Println("Seed default value to visited bundles")
	bundleRes = tx.Model(&models.UserIdentity{}).Where("visited_bundles IS NULL").Update("visited_bundles", []byte(`{}`))
	if bundleRes.Error != nil {
		fmt.Println("Unable to migrate database!", bundleRes.Error.Error())
		tx.Rollback()
		panic(bundleRes.Error)
	}

	fmt.Println("Seed default value to last visited pages")
	visitedRes = tx.Model(&models.UserIdentity{}).Where("last_visited_pages IS NULL").Update("last_visited_pages", []byte(`[]`))
	if bundleRes.Error != nil {
		fmt.Println("Unable to migrate database!", bundleRes.Error.Error())
		tx.Rollback()
		panic(bundleRes.Error)
	}

	dashboardRes = migrateDashboardWidgets(tx)
	if dashboardRes.Error != nil {
		fmt.Println("Unable to migrate database!", dashboardRes.Error.Error())
		tx.Rollback()
		panic(dashboardRes.Error)
	}

	err := tx.Commit().Error

	if err != nil {
		fmt.Println("Unable to migrate database!", err.Error())
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
