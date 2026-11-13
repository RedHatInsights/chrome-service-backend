package e2e

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIRoutes(t *testing.T) {
	config := GetConfig()
	client := &http.Client{Timeout: 30 * time.Second}

	for _, test := range []struct {
		name        string
		path        string
		shouldServe bool
		bodyText    string
	}{
		{
			name:        "Should serve JSON without authentication",
			path:        APIBasePath + "/openapi.json",
			shouldServe: true,
			bodyText:    `"openapi": "3.0.0"`,
		},
		{
			name:        "Should serve YAML without authentication",
			path:        APIBasePath + "/openapi.yaml",
			shouldServe: true,
			bodyText:    "openapi: 3.0.0",
		},
		{
			name: "Should not serve legacy JSON path",
			path: APIBasePath + "/spec/openapi.json",
		},
		{
			name: "Should not serve legacy YAML path",
			path: APIBasePath + "/spec/openapi.yaml",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, config.BaseURL+test.path, nil)
			require.NoError(t, err)

			response, err := client.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			if test.shouldServe {
				assert.Equal(t, http.StatusOK, response.StatusCode)
				assert.Contains(t, string(body), test.bodyText)
				return
			}
			assert.NotEqual(t, http.StatusOK, response.StatusCode)
		})
	}
}
