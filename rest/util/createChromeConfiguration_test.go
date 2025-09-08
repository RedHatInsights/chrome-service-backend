package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseApiSpec(t *testing.T) {
	// Test with valid JSON array
	t.Run("ValidJSON", func(t *testing.T) {
		apiSpecJSON := `[
			{
				"url": "https://console.redhat.com/api/inventory/v1/openapi.json",
				"bundleLabels": ["insights"],
				"frontendName": "inventory-frontend"
			},
			{
				"url": "https://console.redhat.com/api/chrome-service/v1/openapi.json",
				"bundleLabels": ["chrome"],
				"frontendName": "chrome-frontend"
			}
		]`

		result, err := parseApiSpec(apiSpecJSON, "test")
		assert.NoError(t, err)

		// Verify it's valid JSON
		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		assert.NoError(t, err, "Result should be valid JSON")

		// Verify structure
		assert.Len(t, parsed, 2, "Expected 2 API specs")

		// Check first API spec
		firstSpec, ok := parsed[0].(map[string]interface{})
		assert.True(t, ok, "First API spec should be a JSON object")

		assert.Equal(t, "https://console.redhat.com/api/inventory/v1/openapi.json", firstSpec["url"])

		assert.Equal(t, "inventory-frontend", firstSpec["frontendName"])

		// Check bundleLabels
		bundleLabels, ok := firstSpec["bundleLabels"].([]interface{})
		assert.True(t, ok, "bundleLabels should be an array")

		assert.Len(t, bundleLabels, 1)
		assert.Equal(t, "insights", bundleLabels[0])
	})

	// Test with empty string
	t.Run("EmptyString", func(t *testing.T) {
		result, err := parseApiSpec("", "test")
		assert.NoError(t, err)

		// Should return empty array
		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		assert.NoError(t, err, "Result should be valid JSON")

		assert.Empty(t, parsed, "Expected empty array")
	})

	// Test with invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := `[{"url": "https://example.com", "bundleLabels": [`

		_, err := parseApiSpec(invalidJSON, "test")
		assert.Error(t, err, "Expected error for invalid JSON")
	})

	// Test with single API spec
	t.Run("SingleApiSpec", func(t *testing.T) {
		apiSpecJSON := `[{
			"url": "https://console.redhat.com/api/inventory/v1/openapi.json",
			"bundleLabels": ["insights"],
			"frontendName": "test-frontend-service"
		}]`

		result, err := parseApiSpec(apiSpecJSON, "test")
		assert.NoError(t, err)

		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		assert.NoError(t, err, "Result should be valid JSON")

		assert.Len(t, parsed, 1, "Expected 1 API spec")

		spec, ok := parsed[0].(map[string]interface{})
		assert.True(t, ok, "API spec should be a JSON object")

		assert.Equal(t, "test-frontend-service", spec["frontendName"])
	})
}

func TestWriteApiSpecFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Change to temporary directory
	originalWd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")

	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change to temp directory")

	// Ensure we return to original directory
	defer func() {
		os.Chdir(originalWd)
	}()

	// Create static directory
	err = os.MkdirAll("static", 0755)
	assert.NoError(t, err, "Failed to create static directory")

	t.Run("WriteValidApiSpec", func(t *testing.T) {
		apiSpecJSON := `[
			{
				"url": "https://console.redhat.com/api/chrome-service/v1/openapi.json",
				"bundleLabels": ["chrome"],
				"frontendName": "chrome-frontend"
			},
			{
				"url": "https://console.redhat.com/api/inventory/v1/openapi.json",
				"bundleLabels": ["insights"],
				"frontendName": "inventory-frontend"
			}
		]`

		// Parse the API spec
		result, err := parseApiSpec(apiSpecJSON, "test")
		assert.NoError(t, err, "Failed to parse API spec")

		// Write to file
		err = writeConfigFile(result, apiSpecPath)
		assert.NoError(t, err, "Failed to write config file")

		// Verify file was created
		expectedPath := filepath.Join(tempDir, apiSpecPath)
		_, err = os.Stat(expectedPath)
		assert.False(t, os.IsNotExist(err), "Expected file %s was not created", expectedPath)

		// Read and verify content
		fileContent, err := os.ReadFile(expectedPath)
		assert.NoError(t, err, "Failed to read written file")

		// Verify it's valid JSON array
		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		assert.NoError(t, err, "Written file should contain valid JSON")

		// Verify structure
		assert.Len(t, parsedContent, 2, "Expected 2 API specs")

		// Check first spec
		firstSpec, ok := parsedContent[0].(map[string]interface{})
		assert.True(t, ok, "First API spec should be a JSON object")

		assert.Equal(t, "chrome-frontend", firstSpec["frontendName"])

		// Check bundleLabels
		bundleLabels, ok := firstSpec["bundleLabels"].([]interface{})
		assert.True(t, ok, "bundleLabels should be an array")

		assert.Len(t, bundleLabels, 1)
		assert.Equal(t, "chrome", bundleLabels[0])

		// Check URL
		expectedURL := "https://console.redhat.com/api/chrome-service/v1/openapi.json"
		assert.Equal(t, expectedURL, firstSpec["url"])
	})

	t.Run("WriteEmptyApiSpec", func(t *testing.T) {
		// Test writing empty API spec
		result, err := parseApiSpec("", "test")
		assert.NoError(t, err, "Failed to parse empty API spec")

		// Write to a different file to avoid conflicts
		emptyApiSpecPath := "static/api-specs-generated-empty.json"
		err = writeConfigFile(result, emptyApiSpecPath)
		assert.NoError(t, err, "Failed to write empty config file")

		// Verify file was created
		expectedPath := filepath.Join(tempDir, emptyApiSpecPath)
		_, err = os.Stat(expectedPath)
		assert.False(t, os.IsNotExist(err), "Expected file %s was not created", expectedPath)

		// Read and verify content
		fileContent, err := os.ReadFile(expectedPath)
		assert.NoError(t, err, "Failed to read written file")

		// Should be empty JSON array
		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		assert.NoError(t, err, "Written file should contain valid JSON")

		assert.Empty(t, parsedContent, "Expected empty array")
	})
}

func TestApiSpecIntegration(t *testing.T) {
	// Test that simulates the full flow with environment variable
	originalEnv := os.Getenv("FEO_API_SPEC")
	defer os.Setenv("FEO_API_SPEC", originalEnv)

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Change to temporary directory
	originalWd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get working directory")

	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change to temp directory")

	// Ensure we return to original directory
	defer func() {
		os.Chdir(originalWd)
	}()

	// Create static directory
	err = os.MkdirAll("static", 0755)
	assert.NoError(t, err, "Failed to create static directory")

	t.Run("ProcessEnvironmentVariable", func(t *testing.T) {
		// Set test API spec in environment using the exact format from your example
		testApiSpec := `[{
			"url": "https://console.redhat.com/api/inventory/v1/openapi.json",
			"bundleLabels": ["insights"],
			"frontendName": "test-frontend-service"
		}]`

		os.Setenv("FEO_API_SPEC", testApiSpec)

		// Simulate the processing that happens in CreateChromeConfiguration
		apiSpecVar := os.Getenv("FEO_API_SPEC")

		apiSpec, err := parseApiSpec(apiSpecVar, "test")
		assert.NoError(t, err, "Failed to parse API spec from environment")

		err = writeConfigFile(apiSpec, apiSpecPath)
		assert.NoError(t, err, "Failed to write API spec file")

		// Verify the file was created and contains correct data
		expectedPath := filepath.Join(tempDir, apiSpecPath)
		fileContent, err := os.ReadFile(expectedPath)
		assert.NoError(t, err, "Failed to read API spec file")

		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		assert.NoError(t, err, "API spec file should contain valid JSON")

		assert.Len(t, parsedContent, 1, "Expected 1 API spec")

		spec, ok := parsedContent[0].(map[string]interface{})
		assert.True(t, ok, "API spec should be a JSON object")

		assert.Equal(t, "test-frontend-service", spec["frontendName"])

		assert.Equal(t, "https://console.redhat.com/api/inventory/v1/openapi.json", spec["url"])

		// Check bundleLabels
		bundleLabels, ok := spec["bundleLabels"].([]interface{})
		assert.True(t, ok, "bundleLabels should be an array")

		assert.Len(t, bundleLabels, 1)
		assert.Equal(t, "insights", bundleLabels[0])
	})
}
