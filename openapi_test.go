package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupOpenAPIRouter() chi.Router {
	return newRouter(config.Get())
}

func TestOpenAPIRoute(t *testing.T) {
	t.Run("Should serve JSON without authentication through the production router", func(t *testing.T) {
		router := setupOpenAPIRouter()
		request := httptest.NewRequest(http.MethodGet, openAPISpecRoute, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		var document struct {
			OpenAPI string `json:"openapi"`
		}
		assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &document))
		assert.Equal(t, "3.0.0", document.OpenAPI)
	})
}

func TestOpenAPIYAMLRoute(t *testing.T) {
	t.Run("Should serve YAML without authentication through the production router", func(t *testing.T) {
		router := setupOpenAPIRouter()
		request := httptest.NewRequest(http.MethodGet, openAPIYAMLRoute, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), "openapi: 3.0.0")
	})
}

func TestOpenAPIRoutes(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
	}{
		{name: "Should reject legacy JSON path", path: "/api/chrome-service/v1/spec/openapi.json"},
		{name: "Should reject legacy YAML path", path: "/api/chrome-service/v1/spec/openapi.yaml"},
		{name: "Should reject extensionless path", path: openAPIRoute},
		{name: "Should reject unsupported format", path: "/api/chrome-service/v1/openapi.xml"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := setupOpenAPIRouter()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			assert.NotEqual(t, http.StatusOK, response.Code)
		})
	}
}

func TestOpenAPIRouteSupportsHead(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
	}{
		{name: "Should support HEAD for JSON", path: openAPISpecRoute},
		{name: "Should support HEAD for YAML", path: openAPIYAMLRoute},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := setupOpenAPIRouter()
			request := httptest.NewRequest(http.MethodHead, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			assert.Equal(t, http.StatusOK, response.Code)
			assert.Empty(t, response.Body.String())
		})
	}
}
