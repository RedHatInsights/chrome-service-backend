package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
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

// HandlePostLastVisitedPages inserts the most recent pages from chrome. Once they are added,
// all older entries are removed from the table.
func HandlePostLastVisitedPages(recentPages []models.LastVisitedPage, accountId uint) error {
	var ids []uint
	for _, v := range recentPages {
		if err := database.DB.Create(&v).Error; err != nil {
			// If we encounter an error, we want to bail out of the deletion as well. Extra pages will be
			// corrected by a subsequent successful call
			return err
		}
		// Unique ids are given to the entry from Gorm AFTER they are successfully inserted
		ids = append(ids, v.ID)
	}
	// Since we have all the IDs of the newly added pages, we can remove all other pages
	err := database.DB.Where("user_identity_id = ?", accountId).Where("id NOT IN ?", ids).Delete(&models.LastVisitedPage{}).Error
	if err != nil {
		return err
	}
	return nil
}
