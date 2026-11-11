package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func setupOpenAPIRouter() chi.Router {
	router := chi.NewRouter()
	registerOpenAPIRoute(router)
	return router
}

func TestOpenAPIRoute(t *testing.T) {
	router := setupOpenAPIRouter()

	request := httptest.NewRequest("GET", openAPISpecRoute, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Fatalf("expected OpenAPI route to return 200, got %d", response.Code)
	}

	var document struct {
		OpenAPI string `json:"openapi"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatalf("expected OpenAPI route to return JSON: %v", err)
	}
	if document.OpenAPI != "3.0.0" {
		t.Fatalf("expected OpenAPI version 3.0.0, got %q", document.OpenAPI)
	}
}

func TestOpenAPIYAMLRoute(t *testing.T) {
	router := setupOpenAPIRouter()

	request := httptest.NewRequest("GET", openAPIYAMLRoute, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != 200 {
		t.Fatalf("expected OpenAPI YAML route to return 200, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "openapi: 3.0.0") {
		t.Fatal("expected OpenAPI YAML route to return the OpenAPI document")
	}
}

func TestOpenAPIRouteDoesNotExposeOtherSpecFiles(t *testing.T) {
	router := setupOpenAPIRouter()

	for _, path := range []string{
		"/api/chrome-service/v1/spec/openapi.json",
		"/api/chrome-service/v1/spec/openapi.yaml",
	} {
		request := httptest.NewRequest("GET", path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != 404 {
			t.Errorf("expected %s to return 404, got %d", path, response.Code)
		}
	}
}
