package service

import (
  "time"

  "github.com/RedHatInsights/chrome-service-backend/rest/models"
  "github.com/RedHatInsights/chrome-service-backend/rest/database"
)

func GetUserActiveFavoritePages(userID uint)([]models.FavoritePage, error) {
  var activeFavoritePages []models.FavoritePage
  
  err := database.DB.Where("user_identity_id = ?", userID).Where("favorite", true).Find(&activeFavoritePages).Error
  
  return activeFavoritePages, err
}

func GetAllUserFavoritePages(userID uint)([]models.FavoritePage, error) {
  var favoritePages []models.FavoritePage

  err := database.DB.Where("user_identity_id = ?", userID).Find(&favoritePages).Error
  return favoritePages, err
}

func GetUserArchivedFavoritePages(userID uint)([]models.FavoritePage, error) {
  var archivedFavorites []models.FavoritePage

  err := database.DB.
    Where("user_identity_id = ?", userID).
    Where("favorite = ?", false).
    Find(&archivedFavorites).
    Error

    return archivedFavorites, err
}

func CheckIfExistsInDB(allFavoritePages []models.FavoritePage, newFavoritePage models.FavoritePage)(bool) {
  pageExists := false
  
  for _, page := range allFavoritePages {
    if page.Pathname == newFavoritePage.Pathname {
      pageExists = true
      break
    }
  }

  return pageExists
} 

func UpdateFavoritePage(favoritePage models.FavoritePage)(error) {
  var currentPage models.FavoritePage

  err := database.DB.Where("pathname = ?", favoritePage.Pathname).First(&currentPage).Error
  
  if err != nil {
    return err
  }
  
  return database.DB.Model(&currentPage).
    Update("favorite", favoritePage.Favorite).
    Update("updated_at", time.Now()).
    Error
}

func SaveUserFavoritePage(userID uint, newFavoritePage models.FavoritePage)(error) {
  var userFavoritePages []models.FavoritePage

  userFavoritePages, err := GetAllUserFavoritePages(userID)

  if err != nil {
    panic(err)
  }

  alreadyInDB := CheckIfExistsInDB(userFavoritePages, newFavoritePage)

  if alreadyInDB {
    err = UpdateFavoritePage(newFavoritePage)
  } else {
    err = database.DB.Create(&newFavoritePage).Error
  }
 
  return err
}

