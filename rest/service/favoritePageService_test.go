package service

import (
	"errors"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Package-level test constants
const (
	testUserID       uint   = 1
	testAccountID    string = "test-account-123"
	testPathname1    string = "/insights/dashboard"
	testPathname2    string = "/insights/advisor"
	testPathname3    string = "/insights/policies"
)

// setupTestUser initializes the database and creates a test user with automatic cleanup
func setupTestUser(t *testing.T) *models.UserIdentity {
	t.Helper()

	database.Init()

	user := &models.UserIdentity{AccountId: testAccountID}
	err := database.DB.Create(user).Error
	require.NoError(t, err, "unable to create test user")

	// Register cleanup to run after test completes
	t.Cleanup(func() {
		database.DB.Unscoped().Where("user_identity_id = ?", user.ID).Delete(&models.FavoritePage{})
		database.DB.Unscoped().Where("id = ?", user.ID).Delete(&models.UserIdentity{})
	})

	return user
}

// cleanupFavoritePages removes all favorite pages for a user (for use between subtests)
func cleanupFavoritePages(t *testing.T, userID uint) {
	t.Helper()
	database.DB.Unscoped().Where("user_identity_id = ?", userID).Delete(&models.FavoritePage{})
}

// Helper function to create test favorite page
func createTestFavoritePage(t *testing.T, pathname string, favorite bool, userID uint) models.FavoritePage {
	t.Helper()

	return models.FavoritePage{
		Pathname:       pathname,
		Favorite:       favorite,
		UserIdentityID: userID,
	}
}

// Helper function to seed favorite pages in database
func seedFavoritePages(t *testing.T, userID uint, pages []models.FavoritePage) {
	t.Helper()

	for i := range pages {
		pages[i].UserIdentityID = userID
		err := database.DB.Create(&pages[i]).Error
		require.NoError(t, err, "unable to seed favorite page")
	}
}

// TestCheckIfExistsInDB tests the pure function that checks if a page exists
func TestCheckIfExistsInDB(t *testing.T) {
	// Test data
	existingPages := []models.FavoritePage{
		{
			BaseModel:      models.BaseModel{ID: 1},
			Pathname:       testPathname1,
			Favorite:       true,
			UserIdentityID: testUserID,
		},
		{
			BaseModel:      models.BaseModel{ID: 2},
			Pathname:       testPathname2,
			Favorite:       false,
			UserIdentityID: testUserID,
		},
	}

	tests := []struct {
		name            string
		allPages        []models.FavoritePage
		newPage         models.FavoritePage
		expectedExists  bool
		expectedGlobalID uint
	}{
		{
			name:     "page exists in database",
			allPages: existingPages,
			newPage: models.FavoritePage{
				Pathname:       testPathname1,
				Favorite:       true,
				UserIdentityID: testUserID,
			},
			expectedExists:  true,
			expectedGlobalID: 1,
		},
		{
			name:     "page does not exist in database",
			allPages: existingPages,
			newPage: models.FavoritePage{
				Pathname:       testPathname3,
				Favorite:       true,
				UserIdentityID: testUserID,
			},
			expectedExists:  false,
			expectedGlobalID: 0,
		},
		{
			name:     "empty database returns not found",
			allPages: []models.FavoritePage{},
			newPage: models.FavoritePage{
				Pathname:       testPathname1,
				Favorite:       true,
				UserIdentityID: testUserID,
			},
			expectedExists:  false,
			expectedGlobalID: 0,
		},
		{
			name:     "finds archived page with same pathname",
			allPages: existingPages,
			newPage: models.FavoritePage{
				Pathname:       testPathname2,
				Favorite:       true,
				UserIdentityID: testUserID,
			},
			expectedExists:  true,
			expectedGlobalID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, globalID := CheckIfExistsInDB(tt.allPages, tt.newPage)

			assert.Equal(t, tt.expectedExists, exists)
			assert.Equal(t, tt.expectedGlobalID, globalID)
		})
	}
}

// TestGetUserActiveFavoritePages tests retrieving active favorite pages
func TestGetUserActiveFavoritePages(t *testing.T) {
	user := setupTestUser(t)

	// Seed test data - mix of active and archived
	testPages := []models.FavoritePage{
		createTestFavoritePage(t, testPathname1, true, user.ID),
		createTestFavoritePage(t, testPathname2, true, user.ID),
		createTestFavoritePage(t, testPathname3, false, user.ID), // archived
	}
	seedFavoritePages(t, user.ID, testPages)

	// Test retrieval
	activePages, err := GetUserActiveFavoritePages(user.ID)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, activePages, 2, "should return only active favorites")

	// Verify all returned pages are favorites
	for _, page := range activePages {
		assert.True(t, page.Favorite, "all returned pages should have favorite=true")
		assert.Equal(t, user.ID, page.UserIdentityID)
	}
}

// TestGetAllUserFavoritePages tests retrieving all favorite pages for a user
func TestGetAllUserFavoritePages(t *testing.T) {
	user := setupTestUser(t)

	tests := []struct {
		name          string
		seedPages     []models.FavoritePage
		expectedCount int
		validateFunc  func(*testing.T, []models.FavoritePage)
	}{
		{
			name: "returns all pages for user with mixed favorites",
			seedPages: []models.FavoritePage{
				createTestFavoritePage(t, testPathname1, true, user.ID),
				createTestFavoritePage(t, testPathname2, false, user.ID),
				createTestFavoritePage(t, testPathname3, true, user.ID),
			},
			expectedCount: 3,
			validateFunc: func(t *testing.T, pages []models.FavoritePage) {
				t.Helper()
				// Verify it returns BOTH active and archived pages
				activeCount := 0
				archivedCount := 0
				for _, page := range pages {
					if page.Favorite {
						activeCount++
					} else {
						archivedCount++
					}
				}
				assert.Equal(t, 2, activeCount, "should have 2 active favorites")
				assert.Equal(t, 1, archivedCount, "should have 1 archived favorite")
			},
		},
		{
			name:          "returns empty array when user has no pages",
			seedPages:     []models.FavoritePage{},
			expectedCount: 0,
		},
		{
			name: "returns both active and archived pages",
			seedPages: []models.FavoritePage{
				createTestFavoritePage(t, testPathname1, true, user.ID),
				createTestFavoritePage(t, testPathname2, false, user.ID),
			},
			expectedCount: 2,
			validateFunc: func(t *testing.T, pages []models.FavoritePage) {
				t.Helper()
				// Explicitly verify function returns BOTH types
				hasActive := false
				hasArchived := false
				for _, page := range pages {
					if page.Favorite {
						hasActive = true
					} else {
						hasArchived = true
					}
				}
				assert.True(t, hasActive, "should include active favorites")
				assert.True(t, hasArchived, "should include archived pages")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each subtest
			cleanupFavoritePages(t, user.ID)

			// Seed data
			seedFavoritePages(t, user.ID, tt.seedPages)

			// Test
			pages, err := GetAllUserFavoritePages(user.ID)

			// Assertions
			require.NoError(t, err)
			assert.Len(t, pages, tt.expectedCount)

			// Verify all pages belong to the user
			for _, page := range pages {
				assert.Equal(t, user.ID, page.UserIdentityID)
			}

			// Run additional validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, pages)
			}
		})
	}
}

// TestGetUserArchivedFavoritePages tests retrieving archived (favorite=false) pages
func TestGetUserArchivedFavoritePages(t *testing.T) {
	user := setupTestUser(t)

	// Seed test data
	testPages := []models.FavoritePage{
		createTestFavoritePage(t, testPathname1, true, user.ID),  // active
		createTestFavoritePage(t, testPathname2, false, user.ID), // archived
		createTestFavoritePage(t, testPathname3, false, user.ID), // archived
	}
	seedFavoritePages(t, user.ID, testPages)

	// Test retrieval
	archivedPages, err := GetUserArchivedFavoritePages(user.ID)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, archivedPages, 2, "should return only archived pages")

	// Verify all returned pages are archived
	for _, page := range archivedPages {
		assert.False(t, page.Favorite, "all returned pages should have favorite=false")
		assert.Equal(t, user.ID, page.UserIdentityID)
	}
}

// TestDeleteOrUpdateFavoritePage tests the delete and update logic
func TestDeleteOrUpdateFavoritePage(t *testing.T) {
	user := setupTestUser(t)

	tests := []struct {
		name        string
		setupPage   models.FavoritePage
		updatePage  models.FavoritePage
		expectError error
		checkDelete bool
		checkUpdate bool
	}{
		{
			name: "deletes page when favorite is false",
			setupPage: models.FavoritePage{
				Pathname:       testPathname1,
				Favorite:       true,
				UserIdentityID: user.ID,
			},
			updatePage: models.FavoritePage{
				Pathname:       testPathname1,
				Favorite:       false,
				UserIdentityID: user.ID,
			},
			expectError: nil,
			checkDelete: true,
		},
		{
			name: "updates page when favorite is true",
			setupPage: models.FavoritePage{
				Pathname:       testPathname2,
				Favorite:       false,
				UserIdentityID: user.ID,
			},
			updatePage: models.FavoritePage{
				Pathname:       testPathname2,
				Favorite:       true,
				UserIdentityID: user.ID,
			},
			expectError: nil,
			checkUpdate: true,
		},
		{
			name:      "returns error when page does not exist for update",
			setupPage: models.FavoritePage{},
			updatePage: models.FavoritePage{
				Pathname:       "/nonexistent",
				Favorite:       true,
				UserIdentityID: user.ID,
			},
			expectError: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each subtest
			cleanupFavoritePages(t, user.ID)

			// Setup: create page if needed
			if tt.setupPage.Pathname != "" {
				err := database.DB.Create(&tt.setupPage).Error
				require.NoError(t, err, "unable to create setup page")
				// Set ID for delete operation
				tt.updatePage.ID = tt.setupPage.ID
			}

			// Execute
			err := DeleteOrUpdateFavoritePage(tt.updatePage)

			// Assertions
			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectError)
				} else if !errors.Is(err, tt.expectError) {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
			} else {
				require.NoError(t, err)
			}

			// Check if page was deleted
			if tt.checkDelete {
				var count int64
				database.DB.Unscoped().Model(&models.FavoritePage{}).
					Where("pathname = ? AND user_identity_id = ?", tt.updatePage.Pathname, user.ID).
					Count(&count)
				assert.Equal(t, int64(0), count, "page should be deleted")
			}

			// Check if page was updated
			if tt.checkUpdate {
				var updatedPage models.FavoritePage
				err := database.DB.Where("pathname = ?", tt.updatePage.Pathname).
					First(&updatedPage).Error
				require.NoError(t, err)
				assert.True(t, updatedPage.Favorite, "page should be marked as favorite")
			}
		})
	}
}

// TestSaveUserFavoritePage tests the main save function
func TestSaveUserFavoritePage(t *testing.T) {
	user := setupTestUser(t)

	tests := []struct {
		name            string
		existingPages   []models.FavoritePage
		newPage         models.FavoritePage
		expectError     bool
		validateFunc    func(*testing.T, uint)
	}{
		{
			name:          "creates new favorite page",
			existingPages: []models.FavoritePage{},
			newPage:       createTestFavoritePage(t, testPathname1, true, user.ID),
			expectError:   false,
			validateFunc: func(t *testing.T, userID uint) {
				t.Helper()
				var page models.FavoritePage
				err := database.DB.Where("pathname = ? AND user_identity_id = ?",
					testPathname1, userID).First(&page).Error
				require.NoError(t, err)
				assert.True(t, page.Favorite)
			},
		},
		{
			name: "updates existing page to favorite",
			existingPages: []models.FavoritePage{
				createTestFavoritePage(t, testPathname1, false, user.ID),
			},
			newPage:     createTestFavoritePage(t, testPathname1, true, user.ID),
			expectError: false,
			validateFunc: func(t *testing.T, userID uint) {
				t.Helper()
				var page models.FavoritePage
				err := database.DB.Where("pathname = ? AND user_identity_id = ?",
					testPathname1, userID).First(&page).Error
				require.NoError(t, err)
				assert.True(t, page.Favorite)
			},
		},
		{
			name: "deletes page when setting favorite to false",
			existingPages: []models.FavoritePage{
				createTestFavoritePage(t, testPathname1, true, user.ID),
			},
			newPage:     createTestFavoritePage(t, testPathname1, false, user.ID),
			expectError: false,
			validateFunc: func(t *testing.T, userID uint) {
				t.Helper()
				var count int64
				database.DB.Unscoped().Model(&models.FavoritePage{}).
					Where("pathname = ? AND user_identity_id = ?", testPathname1, userID).
					Count(&count)
				assert.Equal(t, int64(0), count, "page should be deleted")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean database before each subtest
			cleanupFavoritePages(t, user.ID)

			// Seed existing pages
			if len(tt.existingPages) > 0 {
				seedFavoritePages(t, user.ID, tt.existingPages)
			}

			// Execute
			err := SaveUserFavoritePage(user.ID, testAccountID, tt.newPage)

			// Assertions
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, user.ID)
				}
			}
		})
	}
}
