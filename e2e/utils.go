package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient wraps an HTTP client with convenience methods for E2E testing
type TestClient struct {
	HTTPClient *http.Client
	Config     *Config
	t          *testing.T
}

// NewTestClient creates a new test client with the given configuration
func NewTestClient(t *testing.T, config *Config) *TestClient {
	return &TestClient{
		HTTPClient: &http.Client{},
		Config:     config,
		t:          t,
	}
}

// XRHIdentity represents the structure of the x-rh-identity header
type XRHIdentity struct {
	Identity Identity `json:"identity"`
}

type Identity struct {
	AccountNumber string `json:"account_number"`
	OrgID         string `json:"org_id"`
	Type          string `json:"type"`
	User          User   `json:"user"`
}

type User struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
	IsOrgAdmin bool  `json:"is_org_admin"`
	IsInternal bool  `json:"is_internal"`
	Locale    string `json:"locale"`
	UserID    string `json:"user_id"`
}

// CreateIdentityHeader creates a base64-encoded x-rh-identity header
func (c *TestClient) CreateIdentityHeader() string {
	identity := XRHIdentity{
		Identity: Identity{
			AccountNumber: c.Config.AccountID,
			OrgID:         c.Config.OrgID,
			Type:          "User",
			User: User{
				Username:   c.Config.Username,
				Email:      fmt.Sprintf("%s@example.com", c.Config.Username),
				FirstName:  "Test",
				LastName:   "User",
				IsActive:   true,
				IsOrgAdmin: false,
				IsInternal: false,
				Locale:     "en_US",
				UserID:     c.Config.UserID,
			},
		},
	}

	jsonBytes, err := json.Marshal(identity)
	require.NoError(c.t, err, "Failed to marshal identity header")

	return base64.StdEncoding.EncodeToString(jsonBytes)
}

// MakeRequest makes an HTTP request with the x-rh-identity header
func (c *TestClient) MakeRequest(method, path string, body interface{}) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.Config.BaseURL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-rh-identity", c.CreateIdentityHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return resp, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, responseBody, nil
}

// GET makes a GET request
func (c *TestClient) GET(path string) (*http.Response, []byte, error) {
	return c.MakeRequest(http.MethodGet, path, nil)
}

// POST makes a POST request
func (c *TestClient) POST(path string, body interface{}) (*http.Response, []byte, error) {
	return c.MakeRequest(http.MethodPost, path, body)
}

// PATCH makes a PATCH request
func (c *TestClient) PATCH(path string, body interface{}) (*http.Response, []byte, error) {
	return c.MakeRequest(http.MethodPatch, path, body)
}

// DELETE makes a DELETE request
func (c *TestClient) DELETE(path string) (*http.Response, []byte, error) {
	return c.MakeRequest(http.MethodDelete, path, nil)
}

// AssertStatusCode asserts that the response has the expected status code
func (c *TestClient) AssertStatusCode(resp *http.Response, expectedStatus int) {
	assert.Equal(c.t, expectedStatus, resp.StatusCode,
		"Expected status code %d but got %d", expectedStatus, resp.StatusCode)
}

// AssertJSONResponse asserts that the response body can be unmarshaled into the given struct
func (c *TestClient) AssertJSONResponse(responseBody []byte, target interface{}) {
	err := json.Unmarshal(responseBody, target)
	require.NoError(c.t, err, "Failed to unmarshal JSON response: %s", string(responseBody))
}

// ListResponse mirrors the util.ListResponse structure
type ListResponse[T any] struct {
	Data []T      `json:"data"`
	Meta ListMeta `json:"meta"`
}

type ListMeta struct {
	Count int `json:"count"`
	Total int `json:"total"`
}

// EntityResponse mirrors the util.EntityResponse structure
type EntityResponse[T any] struct {
	Data T `json:"data"`
}

// ErrorResponse mirrors the util.ErrorResponse structure
type ErrorResponse struct {
	Errors []string `json:"errors"`
}
