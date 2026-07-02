package e2e

import (
	"net/http"
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
			queryParam: "?default=false",
		},
		{
			name:       "Get archived favorite pages",
			queryParam: "?default=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := client.GET("/api/chrome-service/v1/favorite-pages" + tt.queryParam)
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

	resp, body, err := client.POST("/api/chrome-service/v1/favorite-pages", payload)
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
	resp, body, err = client.POST("/api/chrome-service/v1/favorite-pages", payload)
	assert.NoError(t, err, "POST /favorite-pages should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	client.AssertJSONResponse(body, &response)
	assert.NotNil(t, response.Data, "Data should not be nil")
}

func TestSetFavoritePageInvalidRequest(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Send invalid payload (missing required fields)
	payload := map[string]interface{}{
		"invalid": "data",
	}

	resp, _, err := client.POST("/api/chrome-service/v1/favorite-pages", payload)
	assert.NoError(t, err, "POST /favorite-pages should not error")
	client.AssertStatusCode(resp, http.StatusBadRequest)
}
