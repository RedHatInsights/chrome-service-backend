package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	ID                 uint     `json:"id"`
	JobRole            string   `json:"jobRole"`
	ProductsOfInterest []string `json:"productsOfInterest"`
	UserIdentityID     uint     `json:"userIdentityID"`
}

func TestGetUserIdentity(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET(APIBasePath + "/user")
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

	resp, body, err := client.POST(APIBasePath + "/user/update-ui-preview", payload)
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
	resp, body, err = client.POST(APIBasePath + "/user/update-ui-preview", payload)
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

	resp, body, err := client.POST(APIBasePath + "/user/mark-preview-seen", nil)
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

	resp, body, err := client.POST(APIBasePath + "/user/update-active-workspace", payload)
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

	resp, body, err := client.POST(APIBasePath + "/user/visited-bundles", payload)
	assert.NoError(t, err, "POST /user/visited-bundles should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)
	assert.NotNil(t, response.Data, "Response data should not be nil")

	// Verify the bundle is reflected in the user identity
	resp, body, err = client.GET(APIBasePath + "/user")
	assert.NoError(t, err, "GET /user should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var userResponse EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &userResponse)

	visited, exists := userResponse.Data.VisitedBundles["insights"]
	assert.True(t, exists, "insights bundle should exist in visited bundles")
	assert.True(t, visited, "insights bundle should be marked as visited")
}

func TestUserOnboardingFlow(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Step 1: Initial identity fetch - user gets created on first request
	resp, body, err := client.GET(APIBasePath + "/user")
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var identity EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &identity)
	assert.NotEmpty(t, identity.Data.AccountId)

	// Step 2: Enable UI preview and verify via response
	resp, body, err = client.POST(APIBasePath+"/user/update-ui-preview", map[string]bool{"uiPreview": true})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var previewResp EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &previewResp)
	assert.True(t, previewResp.Data.UIPreview, "UIPreview should be enabled after POST")

	// Step 3: Mark preview as seen and verify via response
	resp, body, err = client.POST(APIBasePath+"/user/mark-preview-seen", nil)
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var seenResp EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &seenResp)
	assert.True(t, seenResp.Data.UIPreviewSeen, "UIPreviewSeen should be true after POST")

	// Step 4: Set active workspace and verify via response
	resp, body, err = client.POST(APIBasePath+"/user/update-active-workspace", map[string]string{"activeWorkspace": "onboarding-workspace"})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var workspaceResp EntityResponse[UserIdentityResponse]
	client.AssertJSONResponse(body, &workspaceResp)
	assert.Equal(t, "onboarding-workspace", workspaceResp.Data.ActiveWorkspace)

	// Step 5: Visit bundles and verify each via its POST response
	for _, bundle := range []string{"insights", "openshift", "ansible"} {
		resp, body, err = client.POST(APIBasePath+"/user/visited-bundles", map[string]string{"bundle": bundle})
		assert.NoError(t, err, "POST visited-bundles for %s should not error", bundle)
		client.AssertStatusCode(resp, http.StatusOK)

		var bundleResp EntityResponse[UserIdentityResponse]
		client.AssertJSONResponse(body, &bundleResp)
		visited, exists := bundleResp.Data.VisitedBundles[bundle]
		assert.True(t, exists, "%s bundle should exist in POST response", bundle)
		assert.True(t, visited, "%s bundle should be visited in POST response", bundle)
	}

	// Step 6: Add a favorite page and verify in response
	resp, body, err = client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": "/insights/advisor",
		"favorite": true,
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var favResp ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &favResp)
	found := false
	for _, page := range favResp.Data {
		if page.Pathname == "/insights/advisor" && page.Favorite {
			found = true
			break
		}
	}
	assert.True(t, found, "Favorite page should appear in POST response")

	// Step 7: Store last visited pages and verify in response
	resp, body, err = client.POST(APIBasePath+"/last-visited", map[string][]VisitedPage{
		"pages": {
			{Pathname: "/insights/advisor", Title: "Advisor", Bundle: "insights"},
			{Pathname: "/openshift/overview", Title: "Overview", Bundle: "openshift"},
		},
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var visitedResp ListResponse[VisitedPage]
	client.AssertJSONResponse(body, &visitedResp)
	assert.GreaterOrEqual(t, len(visitedResp.Data), 2, "Should have stored at least 2 visited pages")

	// Cleanup: remove the favorite page
	client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": "/insights/advisor",
		"favorite": false,
	})
}

func TestGetIntercomHash(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET(APIBasePath + "/user/intercom?app=insights")
	assert.NoError(t, err, "GET /user/intercom should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response EntityResponse[map[string]interface{}]
	client.AssertJSONResponse(body, &response)

	// Verify the response contains the expected fields
	// Note: The actual hash value will depend on the server configuration
	assert.NotNil(t, response.Data, "Response data should not be nil")
}
