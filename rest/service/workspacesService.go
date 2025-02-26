package service

import (
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"gorm.io/datatypes"
)

// SaveRecentlyUsedWorkspaces saves a user's recently used workspaces in the database.
func SaveRecentlyUsedWorkspaces(user *models.UserIdentity, recentlyUsedWorkspaces []models.Workspace) error {
	return database.
		DB.
		Model(&user).
		Updates(
			models.UserIdentity{
				ActiveWorkspace:        recentlyUsedWorkspaces[0].Id,
				RecentlyUsedWorkspaces: datatypes.NewJSONType[[]models.Workspace](recentlyUsedWorkspaces),
			},
		).Error
}
