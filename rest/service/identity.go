package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
)

func CreateIdentity(userId string) (models.UserIdentity, error) {
	// var currentUser models.UserIdentity
	identity := models.UserIdentity{
		AccountId: userId,
	}

	err := database.DB.Where("account_id = ?", userId).FirstOrCreate(&identity).Error
	return identity, err
}
