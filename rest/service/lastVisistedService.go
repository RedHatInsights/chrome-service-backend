package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
)

// Get all users last visited pages records
func GetUserslastVisistedPages(accountId uint) ([]models.LastVisitedPage, error) {
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

// Add a new page record to the DB. One user will have 10 record at most
func AddNewPage(pages []models.LastVisitedPage, currentPage models.LastVisitedPage) error {
	var err error
	if len(pages) == util.LAST_VISITED_MAX {
		obsoletePage := pages[len(pages)-1]
		// hard remove from DB. We don't want to keep insanely large history of pages
		err = database.DB.Unscoped().Delete((&obsoletePage)).Error
		if err != nil {
			return err
		}
	} else if len(pages) > util.LAST_VISITED_MAX {
		// if account gets in bad state, remove all until max allowed remain
		for i := util.LAST_VISITED_MAX - 1; i < len(pages); i++ {
			err = database.DB.Unscoped().Delete(pages[i]).Error
			if err != nil {
				return err
			}
		}
	}

	return database.DB.Create(&currentPage).Error
}

func UpdateExistingPage(page models.LastVisitedPage) error {
	return database.DB.Model(&models.LastVisitedPage{}).Where("pathname = ?", page.Pathname).Updates(models.LastVisitedPage{
		Title:    page.Title, // title of a page can change
		Pathname: page.Pathname,
		Bundle:   page.Bundle,
	}).Error
}

func HandlePostLastVisitedPages(accountId uint, currentPage models.LastVisitedPage) error {
	pages, err := GetUserslastVisistedPages(accountId)
	if err != nil {
		return err
	}

	pageExists := CheckExistingPage(pages, currentPage)
	if pageExists {
		err = UpdateExistingPage(currentPage)
	} else {
		err = AddNewPage(pages, currentPage)
	}

	return err
}
