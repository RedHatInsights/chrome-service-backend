package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastVisitedPages(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET(APIBasePath + "/last-visited")
	assert.NoError(t, err, "GET /last-visited should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response ListResponse[VisitedPage]
	client.AssertJSONResponse(body, &response)

	// Verify response structure
	assert.NotNil(t, response.Data, "Data should not be nil")
	assert.GreaterOrEqual(t, response.Meta.Count, 0, "Count should be non-negative")
	assert.Equal(t, response.Meta.Count, len(response.Data), "Count should match data length")
}

func TestStoreLastVisitedPages(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Store a set of last visited pages
	payload := map[string][]VisitedPage{
		"pages": {
			{
				Pathname: "/insights/dashboard",
				Title:    "Insights Dashboard",
				Bundle:   "insights",
			},
			{
				Pathname: "/openshift/overview",
				Title:    "OpenShift Overview",
				Bundle:   "openshift",
			},
		},
	}

	resp, body, err := client.POST(APIBasePath + "/last-visited", payload)
	assert.NoError(t, err, "POST /last-visited should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response ListResponse[VisitedPage]
	client.AssertJSONResponse(body, &response)

	// Verify the pages were stored
	assert.NotNil(t, response.Data, "Data should not be nil")
	assert.GreaterOrEqual(t, len(response.Data), 1, "Should have at least one visited page")

	// Verify the stored pages are in the response
	foundInsights := false
	foundOpenShift := false
	for _, page := range response.Data {
		if page.Pathname == "/insights/dashboard" {
			foundInsights = true
			assert.Equal(t, "Insights Dashboard", page.Title, "Title should match")
			assert.Equal(t, "insights", page.Bundle, "Bundle should match")
		}
		if page.Pathname == "/openshift/overview" {
			foundOpenShift = true
			assert.Equal(t, "OpenShift Overview", page.Title, "Title should match")
			assert.Equal(t, "openshift", page.Bundle, "Bundle should match")
		}
	}
	assert.True(t, foundInsights, "Insights page should be in the response")
	assert.True(t, foundOpenShift, "OpenShift page should be in the response")
}

func TestStoreLastVisitedPagesInvalidRequest(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, _, err := client.MakeRequest(http.MethodPost, APIBasePath + "/last-visited", strings.NewReader("not valid json"))
	assert.NoError(t, err, "POST /last-visited should not error")
	client.AssertStatusCode(resp, http.StatusBadRequest)
}
