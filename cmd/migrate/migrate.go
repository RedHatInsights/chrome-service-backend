package main

import (
	"fmt"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	godotenv.Load()
	database.Init()

	var bundleRes *gorm.DB
	var visitedRes *gorm.DB
	var activeWorkspaceRes *gorm.DB
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
	if visitedRes.Error != nil {
		fmt.Println("Unable to migrate database!", visitedRes.Error.Error())
		tx.Rollback()
		panic(visitedRes.Error)
	}

	fmt.Println("Seed default value to active workspace")
	activeWorkspaceRes = tx.Model(&models.UserIdentity{}).Where("active_workspace IS NULL").Update("active_workspace", []byte(`{}`))
	if activeWorkspaceRes.Error != nil {
		fmt.Println("Unable to migrate database!", activeWorkspaceRes.Error.Error())
		tx.Rollback()
		panic(activeWorkspaceRes.Error)
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

	if activeWorkspaceRes.RowsAffected > 0 {
		logrus.Infof("Migrated %d user identity visited rows", activeWorkspaceRes.RowsAffected)
	}

	logrus.Info("Migration complete")
}
