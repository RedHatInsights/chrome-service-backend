package service

import (
  "time"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	// "github.com/RedHatInsights/chrome-service-backend/rest/util"
)

func GetSelfReport(accountID uint)(models.SelfReport, error) {
  var selfReport models.SelfReport
  
  err := database.DB.Where("user_identity_id = ?", accountID).Find(&selfReport).Error
  return selfReport, err
}

func HandlePatchSelfReport(accountID uint)(models.SelfReport, error) {
  var selfReport models.SelfReport

  err := database.DB.Where("user_identity_id = ?", accountID).Find(&selfReport).Error
  return selfReport, err
}

func HandleNewSelfReport(accountID uint)(models.SelfReport, error) {
  var selfReport models.SelfReport

  err := database.DB.Statement.DB
}
