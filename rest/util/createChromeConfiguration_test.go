package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify it's valid JSON
		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		if err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		// Verify structure
		if len(parsed) != 2 {
			t.Errorf("Expected 2 API specs, got %d", len(parsed))
		}

		// Check first API spec
		firstSpec, ok := parsed[0].(map[string]interface{})
		if !ok {
			t.Fatal("First API spec is not a JSON object")
		}

		if firstSpec["url"] != "https://console.redhat.com/api/inventory/v1/openapi.json" {
			t.Errorf("Expected inventory API URL, got %v", firstSpec["url"])
		}

		if firstSpec["frontendName"] != "inventory-frontend" {
			t.Errorf("Expected frontendName 'inventory-frontend', got %v", firstSpec["frontendName"])
		}

		// Check bundleLabels
		bundleLabels, ok := firstSpec["bundleLabels"].([]interface{})
		if !ok {
			t.Fatal("bundleLabels is not an array")
		}

		if len(bundleLabels) != 1 || bundleLabels[0] != "insights" {
			t.Errorf("Expected bundleLabels ['insights'], got %v", bundleLabels)
		}
	})

	// Test with empty string
	t.Run("EmptyString", func(t *testing.T) {
		result, err := parseApiSpec("", "test")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should return empty array
		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		if err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		if len(parsed) != 0 {
			t.Errorf("Expected empty array, got %v", parsed)
		}
	})

	// Test with invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		invalidJSON := `[{"url": "https://example.com", "bundleLabels": [`

		_, err := parseApiSpec(invalidJSON, "test")
		if err == nil {
			t.Fatal("Expected error for invalid JSON, got none")
		}
	})

	// Test with single API spec
	t.Run("SingleApiSpec", func(t *testing.T) {
		apiSpecJSON := `[{
			"url": "https://console.redhat.com/api/inventory/v1/openapi.json",
			"bundleLabels": ["insights"],
			"frontendName": "test-frontend-service"
		}]`

		result, err := parseApiSpec(apiSpecJSON, "test")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		var parsed []interface{}
		err = json.Unmarshal(result, &parsed)
		if err != nil {
			t.Fatalf("Result is not valid JSON: %v", err)
		}

		if len(parsed) != 1 {
			t.Errorf("Expected 1 API spec, got %d", len(parsed))
		}

		spec, ok := parsed[0].(map[string]interface{})
		if !ok {
			t.Fatal("API spec is not a JSON object")
		}

		if spec["frontendName"] != "test-frontend-service" {
			t.Errorf("Expected frontendName 'test-frontend-service', got %v", spec["frontendName"])
		}
	})
}

func TestWriteApiSpecFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Change to temporary directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Ensure we return to original directory
	defer func() {
		os.Chdir(originalWd)
	}()

	// Create static directory
	err = os.MkdirAll("static", 0755)
	if err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

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
		if err != nil {
			t.Fatalf("Failed to parse API spec: %v", err)
		}

		// Write to file
		err = writeConfigFile(result, apiSpecPath)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Verify file was created
		expectedPath := filepath.Join(tempDir, apiSpecPath)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Fatalf("Expected file %s was not created", expectedPath)
		}

		// Read and verify content
		fileContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		// Verify it's valid JSON array
		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		if err != nil {
			t.Fatalf("Written file contains invalid JSON: %v", err)
		}

		// Verify structure
		if len(parsedContent) != 2 {
			t.Errorf("Expected 2 API specs, got %d", len(parsedContent))
		}

		// Check first spec
		firstSpec, ok := parsedContent[0].(map[string]interface{})
		if !ok {
			t.Fatal("First API spec is not a JSON object")
		}

		if firstSpec["frontendName"] != "chrome-frontend" {
			t.Errorf("Expected frontendName 'chrome-frontend', got %v", firstSpec["frontendName"])
		}

		// Check bundleLabels
		bundleLabels, ok := firstSpec["bundleLabels"].([]interface{})
		if !ok {
			t.Fatal("bundleLabels is not an array")
		}

		if len(bundleLabels) != 1 || bundleLabels[0] != "chrome" {
			t.Errorf("Expected bundleLabels ['chrome'], got %v", bundleLabels)
		}

		// Check URL
		expectedURL := "https://console.redhat.com/api/chrome-service/v1/openapi.json"
		if firstSpec["url"] != expectedURL {
			t.Errorf("Expected URL %s, got %v", expectedURL, firstSpec["url"])
		}
	})

	t.Run("WriteEmptyApiSpec", func(t *testing.T) {
		// Test writing empty API spec
		result, err := parseApiSpec("", "test")
		if err != nil {
			t.Fatalf("Failed to parse empty API spec: %v", err)
		}

		// Write to a different file to avoid conflicts
		emptyApiSpecPath := "static/api-specs-generated-empty.json"
		err = writeConfigFile(result, emptyApiSpecPath)
		if err != nil {
			t.Fatalf("Failed to write empty config file: %v", err)
		}

		// Verify file was created
		expectedPath := filepath.Join(tempDir, emptyApiSpecPath)
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Fatalf("Expected file %s was not created", expectedPath)
		}

		// Read and verify content
		fileContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		// Should be empty JSON array
		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		if err != nil {
			t.Fatalf("Written file contains invalid JSON: %v", err)
		}

		if len(parsedContent) != 0 {
			t.Errorf("Expected empty array, got %v", parsedContent)
		}
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
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	
	// Ensure we return to original directory
	defer func() {
		os.Chdir(originalWd)
	}()

	// Create static directory
	err = os.MkdirAll("static", 0755)
	if err != nil {
		t.Fatalf("Failed to create static directory: %v", err)
	}

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
		if err != nil {
			t.Fatalf("Failed to parse API spec from environment: %v", err)
		}

		err = writeConfigFile(apiSpec, apiSpecPath)
		if err != nil {
			t.Fatalf("Failed to write API spec file: %v", err)
		}

		// Verify the file was created and contains correct data
		expectedPath := filepath.Join(tempDir, apiSpecPath)
		fileContent, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("Failed to read API spec file: %v", err)
		}

		var parsedContent []interface{}
		err = json.Unmarshal(fileContent, &parsedContent)
		if err != nil {
			t.Fatalf("API spec file contains invalid JSON: %v", err)
		}

		if len(parsedContent) != 1 {
			t.Errorf("Expected 1 API spec, got %d", len(parsedContent))
		}

		spec, ok := parsedContent[0].(map[string]interface{})
		if !ok {
			t.Fatal("API spec is not a JSON object")
		}

		if spec["frontendName"] != "test-frontend-service" {
			t.Errorf("Expected frontendName 'test-frontend-service', got %v", spec["frontendName"])
		}

		if spec["url"] != "https://console.redhat.com/api/inventory/v1/openapi.json" {
			t.Errorf("Expected URL 'https://console.redhat.com/api/inventory/v1/openapi.json', got %v", spec["url"])
		}

		// Check bundleLabels
		bundleLabels, ok := spec["bundleLabels"].([]interface{})
		if !ok {
			t.Fatal("bundleLabels is not an array")
		}

		if len(bundleLabels) != 1 || bundleLabels[0] != "insights" {
			t.Errorf("Expected bundleLabels ['insights'], got %v", bundleLabels)
		}
	})
}
