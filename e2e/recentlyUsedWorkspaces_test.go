package e2e

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type Workspace struct {
	Id          string  `json:"id"`
	ParentId    string  `json:"parentId,omitempty"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func TestGetRecentlyUsedWorkspaces(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET("/api/chrome-service/v1/recently-used-workspaces")
	assert.NoError(t, err, "GET /recently-used-workspaces should not error")
	client.AssertStatusCode(resp, http.StatusOK)

	var response ListResponse[Workspace]
	client.AssertJSONResponse(body, &response)

	// Verify response structure
	assert.NotNil(t, response.Data, "Data should not be nil")
	assert.GreaterOrEqual(t, response.Meta.Count, 0, "Count should be non-negative")
	assert.Equal(t, response.Meta.Count, len(response.Data), "Count should match data length")
}

func TestSaveRecentlyUsedWorkspaces(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Create test workspaces
	rootWorkspaceId := uuid.New().String()
	childWorkspaceId := uuid.New().String()
	description := "Test workspace description"

	payload := []Workspace{
		{
			Id:          rootWorkspaceId,
			Type:        "root",
			Name:        "Test Root Workspace",
			Description: &description,
		},
		{
			Id:       childWorkspaceId,
			ParentId: rootWorkspaceId,
			Type:     "standard",
			Name:     "Test Child Workspace",
		},
	}

	resp, body, err := client.POST("/api/chrome-service/v1/recently-used-workspaces", payload)
	assert.NoError(t, err, "POST /recently-used-workspaces should not error")
	client.AssertStatusCode(resp, http.StatusCreated)

	var response ListResponse[Workspace]
	client.AssertJSONResponse(body, &response)

	// Verify the workspaces were stored
	assert.NotNil(t, response.Data, "Data should not be nil")
	assert.GreaterOrEqual(t, len(response.Data), 2, "Should have at least two workspaces")

	// Verify the stored workspaces
	foundRoot := false
	foundChild := false
	for _, workspace := range response.Data {
		if workspace.Id == rootWorkspaceId {
			foundRoot = true
			assert.Equal(t, "root", workspace.Type, "Type should be root")
			assert.Equal(t, "Test Root Workspace", workspace.Name, "Name should match")
		}
		if workspace.Id == childWorkspaceId {
			foundChild = true
			assert.Equal(t, "standard", workspace.Type, "Type should be standard")
			assert.Equal(t, "Test Child Workspace", workspace.Name, "Name should match")
			assert.Equal(t, rootWorkspaceId, workspace.ParentId, "ParentId should match")
		}
	}
	assert.True(t, foundRoot, "Root workspace should be in the response")
	assert.True(t, foundChild, "Child workspace should be in the response")
}

func TestSaveRecentlyUsedWorkspacesValidation(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	tests := []struct {
		name           string
		payload        []Workspace
		expectedErrors []string
	}{
		{
			name:           "Empty payload",
			payload:        []Workspace{},
			expectedErrors: []string{"At least one workspace needs to be specified"},
		},
		{
			name: "Invalid workspace ID",
			payload: []Workspace{
				{
					Id:   "invalid-uuid",
					Type: "root",
					Name: "Test",
				},
			},
			expectedErrors: []string{"Invalid workspace ID"},
		},
		{
			name: "Missing workspace name",
			payload: []Workspace{
				{
					Id:   uuid.New().String(),
					Type: "root",
					Name: "",
				},
			},
			expectedErrors: []string{"workspace's name must not be empty"},
		},
		{
			name: "Invalid workspace type",
			payload: []Workspace{
				{
					Id:   uuid.New().String(),
					Type: "invalid",
					Name: "Test",
				},
			},
			expectedErrors: []string{"Invalid workspace type"},
		},
		{
			name: "Missing parent ID for non-root workspace",
			payload: []Workspace{
				{
					Id:   uuid.New().String(),
					Type: "standard",
					Name: "Test",
				},
			},
			expectedErrors: []string{"parent workspace ID must not be empty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, err := client.POST("/api/chrome-service/v1/recently-used-workspaces", tt.payload)
			assert.NoError(t, err, "POST /recently-used-workspaces should not error")
			client.AssertStatusCode(resp, http.StatusBadRequest)

			var response ErrorResponse
			client.AssertJSONResponse(body, &response)

			// Verify error messages
			assert.NotEmpty(t, response.Errors, "Should have error messages")
			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualError := range response.Errors {
					if assert.Contains(t, actualError, expectedError) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected error '%s' not found in response", expectedError)
			}
		})
	}
}

func TestSaveRecentlyUsedWorkspacesEmptyBody(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.MakeRequest(http.MethodPost, "/api/chrome-service/v1/recently-used-workspaces", nil)
	assert.NoError(t, err, "POST /recently-used-workspaces should not error")
	client.AssertStatusCode(resp, http.StatusBadRequest)

	var response ErrorResponse
	client.AssertJSONResponse(body, &response)

	assert.Contains(t, response.Errors[0], "Request body is empty", "Should have empty body error")
}
