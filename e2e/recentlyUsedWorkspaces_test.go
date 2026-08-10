package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type Workspace struct {
	Id          string  `json:"id"`
	ParentId    string  `json:"parent_id,omitempty"`
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func TestGetRecentlyUsedWorkspaces(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.GET(APIBasePath + "/recently-used-workspaces")
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

	resp, body, err := client.POST(APIBasePath + "/recently-used-workspaces", payload)
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
			resp, body, err := client.POST(APIBasePath + "/recently-used-workspaces", tt.payload)
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

func TestRecentlyUsedWorkspacesLifecycle(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	// Step 1: Save initial workspaces with parent-child relationship
	rootId := uuid.New().String()
	childId := uuid.New().String()
	description := "Root workspace"

	initialPayload := []Workspace{
		{Id: rootId, Type: "root", Name: "Root", Description: &description},
		{Id: childId, ParentId: rootId, Type: "standard", Name: "Child of Root"},
	}

	resp, body, err := client.POST(APIBasePath+"/recently-used-workspaces", initialPayload)
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusCreated)

	var initialResp ListResponse[Workspace]
	client.AssertJSONResponse(body, &initialResp)
	assert.Equal(t, 2, len(initialResp.Data), "Should have 2 workspaces")

	// Verify parent-child relationship persists
	var child Workspace
	for _, ws := range initialResp.Data {
		if ws.Id == childId {
			child = ws
		}
	}
	assert.Equal(t, rootId, child.ParentId, "Child should reference parent")

	// Step 2: Save a new set - replaces previous workspaces
	newId := uuid.New().String()
	replacePayload := []Workspace{
		{Id: newId, Type: "root", Name: "Replacement"},
	}

	resp, body, err = client.POST(APIBasePath+"/recently-used-workspaces", replacePayload)
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusCreated)

	var replaceResp ListResponse[Workspace]
	client.AssertJSONResponse(body, &replaceResp)
	assert.Equal(t, 1, len(replaceResp.Data), "Should have 1 workspace after replacement")
	assert.Equal(t, newId, replaceResp.Data[0].Id)

	// Step 3: Duplicates are deduplicated (most recent first)
	dupId1 := uuid.New().String()
	dupId2 := uuid.New().String()
	dupPayload := []Workspace{
		{Id: dupId1, Type: "root", Name: "First"},
		{Id: dupId2, Type: "root", Name: "Second"},
		{Id: dupId1, Type: "root", Name: "First Again"},
	}

	resp, body, err = client.POST(APIBasePath+"/recently-used-workspaces", dupPayload)
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusCreated)

	var dedupResp ListResponse[Workspace]
	client.AssertJSONResponse(body, &dedupResp)
	assert.Equal(t, 2, len(dedupResp.Data), "Duplicates should be deduplicated to 2 workspaces")

	// Step 4: Exceeding max limit (default 10) trims the list
	var overflowPayload []Workspace
	for i := 0; i < 12; i++ {
		overflowPayload = append(overflowPayload, Workspace{
			Id:   uuid.New().String(),
			Type: "root",
			Name: fmt.Sprintf("Workspace %d", i),
		})
	}

	resp, body, err = client.POST(APIBasePath+"/recently-used-workspaces", overflowPayload)
	assert.NoError(t, err)
	client.AssertStatusCode(resp, http.StatusCreated)

	var overflowResp ListResponse[Workspace]
	client.AssertJSONResponse(body, &overflowResp)
	assert.LessOrEqual(t, len(overflowResp.Data), 10, "Should be trimmed to max limit")
}

func TestSaveRecentlyUsedWorkspacesEmptyBody(t *testing.T) {
	config := GetConfig()
	client := NewTestClient(t, config)

	resp, body, err := client.MakeRequest(http.MethodPost, APIBasePath + "/recently-used-workspaces", nil)
	assert.NoError(t, err, "POST /recently-used-workspaces should not error")
	client.AssertStatusCode(resp, http.StatusBadRequest)

	var response ErrorResponse
	client.AssertJSONResponse(body, &response)

	assert.Contains(t, response.Errors[0], "Request body is empty", "Should have empty body error")
}
