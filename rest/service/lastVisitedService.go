package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"gorm.io/gorm"
)

// GetUsersLastVisitedPages returns all users last visited pages records
func GetUsersLastVisitedPages(accountId uint) ([]models.LastVisitedPage, error) {
	var pages []models.LastVisitedPage
	err := database.DB.Order("updated_at desc").Where("user_identity_id = ?", accountId).Find(&pages).Error
	return pages, err
}

// CheckExistingPage determines if the currently visited page is already in the visited pages array
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

// HandlePostBatchLastVisitedPages inserts the 10 most recent pages from chrome. Once the 10 are added,
// all older entries are removed from the table.
func HandlePostBatchLastVisitedPages(recentPages []models.LastVisitedPage, userId uint) error {
	var ids []uint
	for _, v := range recentPages {
		if err := database.DB.Create(&v).Error; err != nil {
			// If we encounter an error, we want to bail out of the deletion as well. Extra pages will be
			// corrected by a subsequent successful call
			return err
		}
		// ids are given to the entry from Gorm AFTER they are successfully inserted
		ids = append(ids, v.ID)
	}
	// Since we have all the IDs of the newly added pages, we can remove all other pages
	err := database.DB.Where("user_identity_id = ?", userId).Where("id NOT IN ?", ids).Delete(&models.LastVisitedPage{}).Error
	if err != nil {
		return err
	}
	return nil
}

func HandlePostLastVisitedPages(accountId uint, currentPage models.LastVisitedPage) error {
	// Try the entire thing in one transaction
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var pages []models.LastVisitedPage
		// Lock this row until updated
		err := database.DB.Order("updated_at desc").Where("user_identity_id = ?", accountId).Find(&pages).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		pageExists := CheckExistingPage(pages, currentPage)
		if pageExists {
			if err := tx.Model(&models.LastVisitedPage{}).Where("pathname = ?", currentPage.Pathname).Updates(models.LastVisitedPage{
				Title:    currentPage.Title, // title of a page can change
				Pathname: currentPage.Pathname,
				Bundle:   currentPage.Bundle,
			}).Error; err != nil {
				tx.Rollback()
				return err
			}
			return nil
		} else {
			if len(pages) == util.LAST_VISITED_MAX {
				obsoletePage := pages[len(pages)-1]
				// hard remove from DB. We don't want to keep insanely large history of pages
				if err := tx.Unscoped().Delete(&obsoletePage).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else if len(pages) > util.LAST_VISITED_MAX {
				// if account gets in bad state, remove all until only max allowed remain
				for i := util.LAST_VISITED_MAX - 1; i < len(pages); i++ {
					if err := tx.Unscoped().Delete(pages[i]).Error; err != nil {
						tx.Rollback()
						return err
					}
				}
			}
			if err := database.DB.Create(&currentPage).Error; err != nil {
				tx.Rollback()
				return err
			}
			return nil
		}
	})
}
