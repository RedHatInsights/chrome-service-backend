package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// UserIdentityResponse represents the response from GET /user
type UserIdentityResponse struct {
	AccountId        string                 `json:"accountId"`
	FirstLogin       bool                   `json:"firstLogin"`
	DayOne           bool                   `json:"dayOne"`
	LastLogin        string                 `json:"lastLogin"`
	LastVisitedPages []VisitedPage          `json:"lastVisitedPages"`
	FavoritePages    []FavoritePage         `json:"favoritePages"`
	SelfReport       SelfReport             `json:"selfReport"`
	VisitedBundles   map[string]bool        `json:"visitedBundles"`
	UIPreview        bool                   `json:"uiPreview"`
	UIPreviewSeen    bool                   `json:"uiPreviewSeen"`
	ActiveWorkspace  string                 `json:"activeWorkspace"`
}

type VisitedPage struct {
	Pathname string `json:"pathname"`
	Title    string `json:"title"`
	Bundle   string `json:"bundle"`
}

type FavoritePage struct {
	ID       uint   `json:"id"`
	Pathname string `json:"pathname"`
	Favorite bool   `json:"favorite"`
}

type SelfReport struct {
	ID uint `json:"id"`
}

func TestGetUserIdentity(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET("/api/chrome-service/v1/user")
	assert.NoError(t, err, "GET /user should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &response)

	// Verify basic user identity fields are present
	assert.NotEmpty(t, response.Data.AccountId, "AccountId should not be empty")
	assert.NotNil(t, response.Data.VisitedBundles, "VisitedBundles should not be nil")
	assert.NotNil(t, response.Data.LastVisitedPages, "LastVisitedPages should not be nil")
	assert.NotNil(t, response.Data.FavoritePages, "FavoritePages should not be nil")
}

func TestUpdateUserPreview(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Test enabling UI preview
	payload := map[string]bool{
		"uiPreview": true,
	}

	resp, body, err := client.POST("/api/chrome-service/v1/user/update-ui-preview", payload)
	assert.NoError(t, err, "POST /user/update-ui-preview should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Verify the UI preview was updated
	uiPreview, ok := response.Data["uiPreview"].(bool)
	assert.True(t, ok, "uiPreview field should be a boolean")
	assert.True(t, uiPreview, "uiPreview should be true")

	// Test disabling UI preview
	payload["uiPreview"] = false
	resp, body, err = client.POST("/api/chrome-service/v1/user/update-ui-preview", payload)
	assert.NoError(t, err, "POST /user/update-ui-preview should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	client.AssertJSONResponse(body, &response)
	uiPreview, ok = response.Data["uiPreview"].(bool)
	assert.True(t, ok, "uiPreview field should be a boolean")
	assert.False(t, uiPreview, "uiPreview should be false")
}

func TestMarkPreviewSeen(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.POST("/api/chrome-service/v1/user/mark-preview-seen", nil)
	assert.NoError(t, err, "POST /user/mark-preview-seen should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Verify the preview seen flag was updated
	uiPreviewSeen, ok := response.Data["uiPreviewSeen"].(bool)
	assert.True(t, ok, "uiPreviewSeen field should be a boolean")
	assert.True(t, uiPreviewSeen, "uiPreviewSeen should be true")
}

func TestUpdateActiveWorkspace(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	payload := map[string]string{
		"activeWorkspace": "test-workspace-123",
	}

	resp, body, err := client.POST("/api/chrome-service/v1/user/update-active-workspace", payload)
	assert.NoError(t, err, "POST /user/update-active-workspace should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Verify the active workspace was updated
	activeWorkspace, ok := response.Data["activeWorkspace"].(string)
	assert.True(t, ok, "activeWorkspace field should be a string")
	assert.Equal(t, "test-workspace-123", activeWorkspace, "activeWorkspace should match")
}

func TestAddVisitedBundle(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	payload := map[string]string{
		"bundle": "insights",
	}

	resp, body, err := client.POST("/api/chrome-service/v1/user/visited-bundles", payload)
	assert.NoError(t, err, "POST /user/visited-bundles should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Now verify it appears in the visited bundles list
	resp, body, err = client.GET("/api/chrome-service/v1/user/visited-bundles")
	assert.NoError(t, err, "GET /user/visited-bundles should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var bundlesResponse EntityResponse[map[string]bool]
	client.AssertJSONResponse(body, &bundlesResponse)

	// Verify the bundle was added
	visited, exists := bundlesResponse.Data["insights"]
	assert.True(t, exists, "insights bundle should exist in visited bundles")
	assert.True(t, visited, "insights bundle should be marked as visited")
}

func TestGetIntercomHash(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET("/api/chrome-service/v1/user/intercom?app=insights")
	assert.NoError(t, err, "GET /user/intercom should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Verify the response contains the expected fields
	// Note: The actual hash value will depend on the server configuration
	assert.NotNil(t, response.Data, "Response data should not be nil")
}
