package service

import (
	"slices"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
)

// TestSaveRecentlyUsedWorkspaces tests that both the "active workspace" and the workspaces list are saved in the
// database.
func TestSaveRecentlyUsedWorkspaces(t *testing.T) {
	// Create a user fixture in the database.
	user := &models.UserIdentity{
		AccountId: "12345",
	}

	err := database.
		DB.
		Create(&user).
		Error

	if err != nil {
		t.Errorf("unable to save the mock user identity in the database: %s", err)
	}

	// Attempt saving the recently used workspaces for the user that we just created.
	recentlyUsedWorkspaces := []models.Workspace{
		{
			Id: "c93e3bbd-04ac-11f0-ae30-083a885cd988",
		},
		{
			Id: "838191e5-04b4-11f0-8d69-083a885cd988",
		},
		{
			Id: "b6697b96-63ac-4947-b292-018f42f0f55f",
		},
	}

	if err = SaveRecentlyUsedWorkspaces(user, recentlyUsedWorkspaces); err != nil {
		t.Errorf("unable to save the recently used workspaces for the user: %s", err)
	}

	// Retrieve the user again, to avoid using the "user" from above that might get populated by the ORM.
	var retrievedUser *models.UserIdentity
	err = database.
		DB.
		First(&retrievedUser, user.ID).
		Error

	if err != nil {
		t.Errorf("unable to fetch the stored user from the database: %s", err)
	}

	// Assert that both the "active workspace" and the "recently used workspaces" contain the expected values.
	if retrievedUser.ActiveWorkspace != "c93e3bbd-04ac-11f0-ae30-083a885cd988" {
		t.Errorf(`wrong workspace set in the "active workspace" column. Want "%s", got "%s"`, "c93e3bbd-04ac-11f0-ae30-083a885cd988", retrievedUser.ActiveWorkspace)
	}

	if len(recentlyUsedWorkspaces) != len(retrievedUser.RecentlyUsedWorkspaces.Data()) {
		t.Errorf(`unexpected number of recently used workspaces saved. Want "%d", got "%d"`, len(recentlyUsedWorkspaces), len(retrievedUser.RecentlyUsedWorkspaces.Data()))
	}

	for _, got := range retrievedUser.RecentlyUsedWorkspaces.Data() {
		i := slices.IndexFunc(recentlyUsedWorkspaces, func(recentlyUsedWorkspace models.Workspace) bool {
			return recentlyUsedWorkspace == got
		})

		if i == -1 {
			t.Errorf(`recently used workspace "%v" not found in the expected workspaces that should have been saved for the user: %v`, got, recentlyUsedWorkspaces)
		}
	}
}
