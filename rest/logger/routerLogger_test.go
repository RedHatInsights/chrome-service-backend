package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStructuredLogs(t *testing.T, fn func(*bytes.Buffer)) *bytes.Buffer {
	t.Helper()
	standardLogger := logrus.StandardLogger()
	oldOutput, oldLevel, oldFormatter := standardLogger.Out, standardLogger.Level, standardLogger.Formatter
	var output bytes.Buffer
	standardLogger.SetOutput(&output)
	standardLogger.SetLevel(logrus.TraceLevel)
	standardLogger.SetFormatter(&logrus.JSONFormatter{})
	t.Cleanup(func() {
		standardLogger.SetOutput(oldOutput)
		standardLogger.SetLevel(oldLevel)
		standardLogger.SetFormatter(oldFormatter)
	})
	fn(&output)
	return &output
}

func requestLogs(t *testing.T, configuredLevel string, status int) *bytes.Buffer {
	t.Helper()
	return captureStructuredLogs(t, func(_ *bytes.Buffer) {
		router := chi.NewRouter()
		router.Use(middleware.RequestID)
		router.Use(InjectLogger)
		router.Use(middleware.RequestLogger(NewLogger(&config.ChromeServiceConfig{LogLevel: configuredLevel}, logrus.StandardLogger())))
		router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), util.IDENTITY_CTX_KEY, &identity.XRHID{
				Identity: identity.Identity{
					OrgID:         "org-123",
					AccountNumber: "account-456",
					User:          &identity.User{UserID: "user-789"},
				},
			})
			EnrichLoggerWithIdentity(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			})).ServeHTTP(w, r.WithContext(ctx))
		})
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/users/secret-value?token=query-secret", nil)
		request.Header.Set("Authorization", "bearer-header-secret")
		request.Header.Set("X-RH-Identity", "identity-header-secret")
		router.ServeHTTP(response, request)
		assert.Equal(t, status, response.Code)
	})
}

func TestCompletionSeverityFilteringAndCorrelation(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		status    int
		wantLevel string
	}{
		{name: "error level filters success", level: "error", status: http.StatusOK},
		{name: "error level filters client errors", level: "error", status: http.StatusBadRequest},
		{name: "error level logs server errors", level: "error", status: http.StatusInternalServerError, wantLevel: "error"},
		{name: "warn level logs client errors", level: "warn", status: http.StatusBadRequest, wantLevel: "warning"},
		{name: "warn level logs server errors", level: "warn", status: http.StatusInternalServerError, wantLevel: "error"},
		{name: "warn level filters success", level: "warn", status: http.StatusOK},
		{name: "info level logs success", level: "info", status: http.StatusOK, wantLevel: "info"},
		{name: "info level logs client errors", level: "info", status: http.StatusBadRequest, wantLevel: "warning"},
		{name: "info level logs server errors", level: "info", status: http.StatusInternalServerError, wantLevel: "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := requestLogs(t, tt.level, tt.status)
			if tt.wantLevel == "" {
				assert.Empty(t, strings.TrimSpace(output.String()))
				return
			}

			var entry map[string]any
			require.NoError(t, json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry))
			assert.Equal(t, tt.wantLevel, entry["level"])
			assert.Equal(t, float64(tt.status), entry["status"])
			assert.Equal(t, "/users/{id}", entry["route"])
			assert.Equal(t, "user-789", entry["user_id"])
			assert.Equal(t, "org-123", entry["org_id"])
			assert.NotEmpty(t, entry["request_id"])
			assert.NotContains(t, output.String(), "secret-value")
			assert.NotContains(t, output.String(), "query-secret")
			assert.NotContains(t, output.String(), "bearer-header-secret")
			assert.NotContains(t, output.String(), "identity-header-secret")
			assert.NotContains(t, output.String(), "remote_addr")
		})
	}
}

func TestRecovererCorrelatesWithoutLoggingPanicValue(t *testing.T) {
	output := captureStructuredLogs(t, func(_ *bytes.Buffer) {
		router := chi.NewRouter()
		router.Use(middleware.RequestID)
		router.Use(InjectLogger)
		router.Use(middleware.RequestLogger(NewLogger(&config.ChromeServiceConfig{LogLevel: "error"}, logrus.StandardLogger())))
		router.Use(Recoverer)
		router.Get("/panic/{id}", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), util.IDENTITY_CTX_KEY, &identity.XRHID{
				Identity: identity.Identity{
					OrgID: "org-panic",
					User:  &identity.User{UserID: "user-panic"},
				},
			})
			EnrichLoggerWithIdentity(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				panic("panic-secret-value")
			})).ServeHTTP(w, r.WithContext(ctx))
		})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic/secret-id", nil))
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 2, "expected recovered panic event and HTTP completion event")
	var recovered, completed map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &recovered))
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &completed))
	assert.Equal(t, "request panic recovered", recovered["msg"])
	assert.Equal(t, "request completed", completed["msg"])
	assert.Equal(t, "user-panic", recovered["user_id"])
	assert.Equal(t, "org-panic", recovered["org_id"])
	assert.Equal(t, "error", completed["level"])
	assert.Equal(t, float64(http.StatusInternalServerError), completed["status"])
	assert.Equal(t, recovered["request_id"], completed["request_id"])
	assert.NotContains(t, output.String(), "panic-secret-value")
	assert.NotContains(t, output.String(), "secret-id")
}

func TestSafeCorrelationValueBoundsAndEscapes(t *testing.T) {
	value := safeCorrelationValue("user\n"+strings.Repeat("x", 200), 128)
	assert.NotContains(t, value, "\n")
	assert.LessOrEqual(t, len(value), 128)
	assert.LessOrEqual(t, len(safeCorrelationValue(strings.Repeat("é", 100), 128)), 128)
}
