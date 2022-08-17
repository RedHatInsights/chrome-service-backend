package service

import (
 "github.com/RedHatInsights/chrome-service-backend/rest/models"
 "github.com/RedHatInsights/chrome-service-backend/rest/database"
)

func GetUserActiveFavoritePages(userID uint)([]models.FavoritePage, error) {
  var activeFavoritePages []models.FavoritePage
  
  err := database.DB.Where("user_identity_id ?", userID).Where("favorite", true).Find(&activeFavoritePages).Error
  
  return activeFavoritePages, err
}
