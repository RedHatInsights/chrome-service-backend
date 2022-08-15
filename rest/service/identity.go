package service

import (
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
)

func GetUserIdentityByID(userID uint) (models.UserIdentity, error) {
	var user models.UserIdentity
	err := database.DB.First(&user, userID).Error
	
	return user, err
}

func CreateIdentity(userId string) (models.UserIdentity, error) {
	identity := models.UserIdentity{
		AccountId: userId,
		FirstLogin: true,
		DayOne: true,
		LastLogin: time.Now(),
	}

	err := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity).Error
	return identity, err
}
