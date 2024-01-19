package service

import (
	"fmt"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/sirupsen/logrus"
)

func GetUserActiveFavoritePages(userID uint) ([]models.FavoritePage, error) {
	var activeFavoritePages []models.FavoritePage

	err := database.DB.Where("user_identity_id = ?", userID).Where("favorite", true).Find(&activeFavoritePages).Error

	return activeFavoritePages, err
}

func GetAllUserFavoritePages(userID uint) ([]models.FavoritePage, error) {
	var favoritePages []models.FavoritePage

	err := database.DB.Where("user_identity_id = ?", userID).Find(&favoritePages).Error
	return favoritePages, err
}

func GetUserArchivedFavoritePages(userID uint) ([]models.FavoritePage, error) {
	var archivedFavorites []models.FavoritePage

	err := database.DB.
		Where("user_identity_id = ?", userID).
		Where("favorite = ?", false).
		Find(&archivedFavorites).
		Error

	return archivedFavorites, err
}

func CheckIfExistsInDB(allFavoritePages []models.FavoritePage, newFavoritePage models.FavoritePage) bool {
	pageExists := false

	for _, page := range allFavoritePages {
		if page.Pathname == newFavoritePage.Pathname {
			pageExists = true
			break
		}
	}

	return pageExists
}

func UpdateFavoritePage(favoritePage models.FavoritePage) error {
	return database.DB.Model(&models.FavoritePage{}).Where("pathname = ?", favoritePage.Pathname).Update("favorite", favoritePage.Favorite).Error
}

func debugFavoritesEntry(accountId string, payload models.FavoritePage) {
	c := config.Get()
	for _, i := range c.DebugConfig.DebugFavoriteIds {
		if i == accountId {
			logrus.Warningln(fmt.Sprintf("\n_____\nDEBUG_FAVORITES_ACCOUNT_ID: %s\nDEBUG_FAVORITES_PATH: %s\nDEBUG_FAVORITES_FLAG: %s\n_____", accountId, payload.Pathname, fmt.Sprint(payload.Favorite)))
		}
	}
}

func SaveUserFavoritePage(userID uint, accountId string, newFavoritePage models.FavoritePage) error {
	var userFavoritePages []models.FavoritePage

	userFavoritePages, err := GetAllUserFavoritePages(userID)

	if err != nil {
		panic(err)
	}

	alreadyInDB := CheckIfExistsInDB(userFavoritePages, newFavoritePage)

	if alreadyInDB {
		err = UpdateFavoritePage(newFavoritePage)
		debugFavoritesEntry(accountId, newFavoritePage)
	} else {
		debugFavoritesEntry(accountId, newFavoritePage)
		err = database.DB.Create(&newFavoritePage).Error
	}

	return err
}
