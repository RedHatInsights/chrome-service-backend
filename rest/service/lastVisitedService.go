package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Get all users last visited pages records
func GetUsersLastVisitedPages(accountId uint) ([]models.LastVisitedPage, error) {
	var pages []models.LastVisitedPage
	err := database.DB.Order("updated_at desc").Where("user_identity_id = ?", accountId).Find(&pages).Error
	return pages, err
}

// Check if the currently visited page is already in the visited pages array
func CheckExistingPage(pages []models.LastVisitedPage, currentPage models.LastVisitedPage) bool {
	pageExists := false
	for _, page := range pages {
		if page.Pathname == currentPage.Pathname {
			pageExists = true
			break
		}
	}
	return pageExists
}

func HandlePostLastVisitedPages(accountId uint, currentPage models.LastVisitedPage) error {
	// Try the entire thing in one transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var pages []models.LastVisitedPage
		// Lock this row until updated
		err := database.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Order("updated_at desc").Where("user_identity_id = ?", accountId).Find(&pages).Error
		if err != nil {
			return err
		}
		pageExists := CheckExistingPage(pages, currentPage)
		if pageExists {
			if err := tx.Model(&models.LastVisitedPage{}).Where("pathname = ?", currentPage.Pathname).Updates(models.LastVisitedPage{
				Title:    currentPage.Title, // title of a page can change
				Pathname: currentPage.Pathname,
				Bundle:   currentPage.Bundle,
			}).Error; err != nil {
				return err
			}
			return nil
		} else {
			if len(pages) == util.LAST_VISITED_MAX {
				obsoletePage := pages[len(pages)-1]
				// hard remove from DB. We don't want to keep insanely large history of pages
				if err := tx.Unscoped().Delete(&obsoletePage).Error; err != nil {
					return err
				}
			} else if len(pages) > util.LAST_VISITED_MAX {
				// if account gets in bad state, remove all until only max allowed remain
				for i := util.LAST_VISITED_MAX - 1; i < len(pages); i++ {
					if err := tx.Unscoped().Delete(pages[i]).Error; err != nil {
						return err
					}
				}
			}
			if err := database.DB.Create(&currentPage).Error; err != nil {
				return err
			}
			return nil
		}
	})
}
