package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFavoritePages(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	tests := []struct {
		name       string
		queryParam string
	}{
		{
			name:       "Get all favorite pages",
			queryParam: "?getAll=true",
		},
		{
			name:       "Get active favorite pages",
			queryParam: "?archived=false",
		},
		{
			name:       "Get archived favorite pages",
			queryParam: "?archived=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := client.GET(APIBasePath + "/favorite-pages" + tt.queryParam)
			assert.NoError(t, err, "GET /favorite-pages should not error")
			client.AssertStatusCode(resp, http.StatusOK)

			var response ListResponse[FavoritePage]
			client.AssertJSONResponse(body, &response)

			// Verify response structure
			assert.NotNil(t, response.Data, "Data should not be nil")
			assert.GreaterOrEqual(t, response.Meta.Count, 0, "Count should be non-negative")
			assert.Equal(t, response.Meta.Count, len(response.Data), "Count should match data length")
		})
	}
}

func TestSetFavoritePage(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Add a favorite page
	payload := map[string]interface{}{
		"pathname": "/insights/dashboard",
		"favorite": true,
	}

	resp, body, err := client.POST(APIBasePath + "/favorite-pages", payload)
	assert.NoError(t, err, "POST /favorite-pages should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &response)

	// Verify the favorite page was added
	assert.NotNil(t, response.Data, "Data should not be nil")
	assert.GreaterOrEqual(t, len(response.Data), 1, "Should have at least one favorite page")

	// Verify the added page exists in the list
	found := false
	for _, page := range response.Data {
		if page.Pathname == "/insights/dashboard" && page.Favorite {
			found = true
			break
		}
	}
	assert.True(t, found, "Added favorite page should be in the response")

	// Remove the favorite page
	payload["favorite"] = false
	resp, body, err = client.POST(APIBasePath + "/favorite-pages", payload)
	assert.NoError(t, err, "POST /favorite-pages should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	client.AssertJSONResponse(body, &response)
	assert.NotNil(t, response.Data, "Data should not be nil")
}

func TestFavoritePageLifecycle(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	pagePath1 := "/settings/notifications"
	pagePath2 := "/openshift/clusters"

	// Step 1: Add two favorite pages, verify each via POST response
	resp, body, err := client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": pagePath1,
		"favorite": true,
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var addResp1 ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &addResp1)
	foundPage1 := false
	for _, page := range addResp1.Data {
		if page.Pathname == pagePath1 && page.Favorite {
			foundPage1 = true
		}
	}
	assert.True(t, foundPage1, "First page should be in POST response")

	resp, body, err = client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": pagePath2,
		"favorite": true,
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	// Step 3: Second POST response should contain both active favorites
	var addResp2 ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &addResp2)

	foundPage1 = false
	foundPage2 := false
	for _, page := range addResp2.Data {
		if page.Pathname == pagePath1 {
			foundPage1 = true
		}
		if page.Pathname == pagePath2 {
			foundPage2 = true
		}
	}
	assert.True(t, foundPage1, "First page should be in active favorites")
	assert.True(t, foundPage2, "Second page should be in active favorites")

	// Step 4: Archive one page (set favorite=false)
	resp, body, err = client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": pagePath1,
		"favorite": false,
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	// Step 5: Verify archived page is gone from active list but second page remains
	var afterArchiveResp ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &afterArchiveResp)

	for _, page := range afterArchiveResp.Data {
		assert.NotEqual(t, pagePath1, page.Pathname, "Archived page should not be in active list")
	}

	stillActive := false
	for _, page := range afterArchiveResp.Data {
		if page.Pathname == pagePath2 {
			stillActive = true
		}
	}
	assert.True(t, stillActive, "Second page should still be active")

	// Step 6: Re-favorite the archived page
	resp, body, err = client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
		"pathname": pagePath1,
		"favorite": true,
	})
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusOK)

	var reFavResp ListResponse[FavoritePage]
	client.AssertJSONResponse(body, &reFavResp)

	reAdded := false
	for _, page := range reFavResp.Data {
		if page.Pathname == pagePath1 && page.Favorite {
			reAdded = true
		}
	}
	assert.True(t, reAdded, "Re-favorited page should appear as active")

	// Cleanup: remove both test pages
	for _, path := range []string{pagePath1, pagePath2} {
		client.POST(APIBasePath+"/favorite-pages", map[string]interface{}{
			"pathname": path,
			"favorite": false,
		})
	}
}

func TestSetFavoritePageInvalidRequest(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, _, err := client.MakeRequest(http.MethodPost, APIBasePath + "/favorite-pages", strings.NewReader("not valid json"))
	assert.NoError(t, err, "POST /favorite-pages should not error")
	client.AssertStatusCode(resp, http.StatusBadRequest)
}
