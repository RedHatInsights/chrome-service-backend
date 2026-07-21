package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserSelfReport(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET("/api/chrome-service/v1/self-report")
	assert.NoError(t, err, "GET /self-report should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response SelfReport
	client.AssertJSONResponse(body, &response)

	// Verify response structure - ID should be present
	assert.NotZero(t, response.ID, "ID should not be zero")
}

func TestUpdateUserSelfReport(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// First, get the current self report to ensure it exists
	resp, body, err := client.GET("/api/chrome-service/v1/self-report")
	assert.NoError(t, err, "GET /self-report should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var currentReport SelfReport
	client.AssertJSONResponse(body, &currentReport)

	// Update the self report with some test data
	// Note: The exact fields available depend on the SelfReport model
	// This is a minimal test that verifies the endpoint works
	payload := map[string]interface{}{
		"id": currentReport.ID,
	}

	resp, body, err = client.PATCH("/api/chrome-service/v1/self-report", payload)
	assert.NoError(t, err, "PATCH /self-report should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var updatedReport SelfReport
	client.AssertJSONResponse(body, &updatedReport)

	// Verify the report was updated (ID should remain the same)
	assert.Equal(t, currentReport.ID, updatedReport.ID, "ID should not change")
}
