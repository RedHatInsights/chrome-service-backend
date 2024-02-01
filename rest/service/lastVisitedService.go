package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
)

// GetUsersLastVisitedPages returns all users last visited pages records
func GetUsersLastVisitedPages(accountId uint) ([]models.LastVisitedPage, error) {
	var pages []models.LastVisitedPage
	err := database.DB.Order("updated_at desc").Where("user_identity_id = ?", accountId).Find(&pages).Error
	return pages, err
}

// HandlePostLastVisitedPages inserts the most recent pages from chrome. Once they are added,
// all older entries are removed from the table.
func HandlePostLastVisitedPages(recentPages []models.LastVisitedPage, accountId uint) error {
	// change the ID in place to avoid a user with ID 0
	for k, _ := range recentPages {
		recentPages[k].UserIdentityID = accountId
	}
	if err := database.DB.Create(&recentPages).Error; err != nil {
		logrus.Debugf("Pages to be inserted: %+v\n", recentPages)
		// If we encounter an error, we want to bail out of the deletion as well. Extra pages will be
		// corrected by a subsequent successful call
		return err
	}
	var ids []uint
	// Unique ids are given to the entry from Gorm AFTER they are successfully inserted
	for _, v := range recentPages {
		ids = append(ids, v.ID)
	}
	logrus.Debugf("Pages to be inserted: %+v\n", recentPages)
	logrus.Debugf("Inserted IDs: %+v\n", ids)
	// Since we have all the IDs of the newly added pages, we can remove all other pages
	err := database.DB.Unscoped().Where("user_identity_id = ?", accountId).Where("id NOT IN ?", ids).Delete(&models.LastVisitedPage{}).Error
	if err != nil {
		return err
	}
	return nil
}
