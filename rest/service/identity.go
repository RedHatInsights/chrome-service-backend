package service

import (
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
)

// Get user data complete with it's related tables.
func GetUserIdentityData(user models.UserIdentity) (models.UserIdentity, error) {
	var lastVisitedPages []models.LastVisitedPage

	err := database.DB.Model(&user).Association("LastVisitedPages").Find(&lastVisitedPages)
	user.LastVisitedPages = lastVisitedPages
	return user, err
}

// Create the user object and add the row if not already in DB
func CreateIdentity(userId string) (models.UserIdentity, error) {
	identity := models.UserIdentity{
		AccountId:        userId,
		FirstLogin:       true,
		DayOne:           true,
		LastLogin:        time.Now(),
		LastVisitedPages: []models.LastVisitedPage{},
		FavoritePages:    []models.FavoritePage{},
		SelfReport:       models.SelfReport{},
	}

	res := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity)
	err := res.Error

	return identity, err
}
