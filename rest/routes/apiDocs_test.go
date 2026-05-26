package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withTempDir(t *testing.T, fn func()) {
	t.Helper()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}
	fn()
}

func writeSpecsFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll("static", 0o755); err != nil {
		t.Fatalf("failed to create static directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join("static", "specs-generated.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write specs file: %v", err)
	}
}

func TestGetApiDocsSuccess(t *testing.T) {
	withTempDir(t, func() {
		expected := `{"notifications":[{"url":"https://example.com/api/v1/openapi.json","bundleLabels":["settings"],"spec":{"openapi":"3.0.0"}}]}`
		writeSpecsFile(t, expected)

		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, expected, recorder.Body.String())
	})
}

func TestGetApiDocsFileNotFound(t *testing.T) {
	withTempDir(t, func() {
		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "{}", recorder.Body.String())
	})
}

func TestGetApiDocsContentType(t *testing.T) {
	withTempDir(t, func() {
		writeSpecsFile(t, `{"rbac":[]}`)

		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	})
}

func TestGetApiDocsReadError(t *testing.T) {
	// chmod 0o000 has no effect when running as root (e.g. in CI containers)
	if os.Getuid() == 0 {
		t.Skip("skipping permission test when running as root")
	}

	withTempDir(t, func() {
		writeSpecsFile(t, "test")
		os.Chmod(filepath.Join("static", "specs-generated.json"), 0o000)

		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)

		var errResp map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		_, hasErrors := errResp["errors"]
		assert.True(t, hasErrors, "response should contain errors field")
	})
}

func TestGetApiDocsBundleFilter(t *testing.T) {
	withTempDir(t, func() {
		specsJSON := `{
			"notifications":[{"url":"https://example.com/notifications/v1","bundleLabels":["settings"],"spec":{}}],
			"rbac":[{"url":"https://example.com/rbac/v1","bundleLabels":["iam"],"spec":{}}],
			"sources":[
				{"url":"https://example.com/integrations/v1","bundleLabels":["settings"],"spec":{}},
				{"url":"https://example.com/sources/v3","bundleLabels":["insights"],"spec":{}}
			]
		}`
		writeSpecsFile(t, specsJSON)

		request, _ := http.NewRequest("GET", "/api-docs?bundle=settings", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var result map[string][]apiDocEntry
		err := json.Unmarshal(recorder.Body.Bytes(), &result)
		assert.NoError(t, err)

		assert.Len(t, result, 2, "should return only frontends with settings bundle")
		assert.Len(t, result["notifications"], 1)
		assert.Len(t, result["sources"], 1)
		assert.Equal(t, "https://example.com/integrations/v1", result["sources"][0].URL)
		_, hasRbac := result["rbac"]
		assert.False(t, hasRbac, "rbac should be excluded (iam bundle only)")
	})
}

func TestGetApiDocsBundleFilterNoMatch(t *testing.T) {
	withTempDir(t, func() {
		writeSpecsFile(t, `{"notifications":[{"url":"https://example.com","bundleLabels":["settings"],"spec":{}}]}`)

		request, _ := http.NewRequest("GET", "/api-docs?bundle=nonexistent", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, `{}`, recorder.Body.String())
	})
}

func TestGetApiDocsCacheControlHeader(t *testing.T) {
	withTempDir(t, func() {
		writeSpecsFile(t, `{}`)

		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		assert.Equal(t, "max-age=3600", recorder.Header().Get("Cache-Control"))
	})
}

func TestGetApiDocsNoResponseWrapper(t *testing.T) {
	withTempDir(t, func() {
		writeSpecsFile(t, `{"svc":[{"url":"https://example.com","bundleLabels":["test"],"spec":{}}]}`)

		request, _ := http.NewRequest("GET", "/api-docs", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(GetApiDocs)
		handler.ServeHTTP(recorder, request)

		var result map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &result)
		assert.NoError(t, err)

		_, hasData := result["data"]
		assert.False(t, hasData, "response should not be wrapped in {\"data\": ...}")

		_, hasSvc := result["svc"]
		assert.True(t, hasSvc, "response should contain raw file content")
	})
}
