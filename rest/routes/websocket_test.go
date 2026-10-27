package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	requestlogger "github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCheckOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{"console.redhat.com", "https://console.redhat.com", true},
		{"console.stage.redhat.com", "https://console.stage.redhat.com", true},
		{"console.dev.redhat.com", "https://console.dev.redhat.com", true},
		{"subdomain of console.redhat.com", "https://beta.console.redhat.com", true},
		{"subdomain of console.stage.redhat.com", "https://beta.console.stage.redhat.com", true},
		{"subdomain of console.dev.redhat.com", "https://beta.console.dev.redhat.com", true},
		{"local dev proxy stage", "https://stage.foo.redhat.com:1337", true},
		{"local dev proxy prod", "https://prod.foo.redhat.com:1337", true},
		{"foo.redhat.com subdomain no port", "https://stage.foo.redhat.com", true},

		{"empty origin", "", false},
		{"http scheme rejected", "http://console.redhat.com", false},
		{"ws scheme rejected", "ws://console.redhat.com", false},
		{"wss scheme rejected", "wss://console.redhat.com", false},
		{"unrelated domain", "https://evil.example.com", false},
		{"suffix spoof", "https://evilconsole.redhat.com", false},
		{"suffix spoof stage", "https://notconsole.stage.redhat.com", false},
		{"malformed URL", "://not-a-url", false},
		{"bare hostname no scheme", "console.redhat.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/ws", nil)
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			assert.Equal(t, tt.want, checkOrigin(r))
		})
	}
}

func TestInspectOriginHeaderDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		set        func(http.Header)
		wantState  string
		wantReason string
		wantRepr   string
	}{
		{"missing", func(http.Header) {}, "missing", "missing", "<missing>"},
		{"null", func(h http.Header) { h["Origin"] = nil }, "null", "missing", "<null>"},
		{"empty", func(h http.Header) { h["Origin"] = []string{""} }, "empty", "missing", "<empty>"},
		{"malformed", func(h http.Header) { h.Set("Origin", "://bad") }, "present", "malformed", "<malformed>"},
		{"non HTTPS", func(h http.Header) { h.Set("Origin", "http://console.redhat.com") }, "present", "disallowed_scheme", "http://<allowed-host>"},
		{"unlisted", func(h http.Header) { h.Set("Origin", "https://evil.example/path?token=private#fragment") }, "present", "disallowed_host", "https://<unlisted-host>"},
		{"userinfo removed", func(h http.Header) { h.Set("Origin", "https://private:secret@evil.example/?password=secret#private") }, "present", "disallowed_host", "https://<unlisted-host>"},
		{"valid accepted", func(h http.Header) { h.Set("Origin", "https://console.redhat.com") }, "present", "", "https://<allowed-host>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ws", nil)
			tt.set(r.Header)
			info := inspectOriginHeader(r)
			assert.Equal(t, tt.wantState, info.state)
			assert.Equal(t, tt.wantReason, info.rejectionReason)
			assert.Equal(t, tt.wantRepr, info.representation)
			assert.Equal(t, tt.wantReason == "", checkOrigin(r))
		})
	}
}

func TestHandleWsConnectionLogsStructuredOriginRejection(t *testing.T) {
	tests := []struct {
		name       string
		setOrigin  func(http.Header)
		wantReason string
		wantState  string
	}{
		{"missing", func(http.Header) {}, "missing", "missing"},
		{"null", func(h http.Header) { h["Origin"] = nil }, "missing", "null"},
		{"malformed", func(h http.Header) { h.Set("Origin", "://bad") }, "malformed", "present"},
		{"non HTTPS", func(h http.Header) { h.Set("Origin", "http://console.redhat.com") }, "disallowed_scheme", "present"},
		{"unlisted", func(h http.Header) { h.Set("Origin", "https://evil.example") }, "disallowed_host", "present"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, output := newTestRequestLogger("request-correlation-123")
			r := websocketUpgradeRequest()
			tt.setOrigin(r.Header)
			r = r.WithContext(context.WithValue(r.Context(), util.IDENTITY_CTX_KEY, &identity.XRHID{Identity: identity.Identity{User: &identity.User{}}}))
			r = r.WithContext(requestlogger.WithLogger(context.WithValue(r.Context(), middleware.RequestIDKey, "request-correlation-123"), entry))
			w := httptest.NewRecorder()

			HandleWsConnection(w, r)

			assert.Equal(t, http.StatusForbidden, w.Code)
			var event map[string]interface{}
			assert.NoError(t, json.Unmarshal([]byte(output.String()), &event))
			assert.Equal(t, 1, strings.Count(output.String(), "\n"), "rejection should emit one diagnostic event")
			assert.Equal(t, "websocket_origin_rejected", event["event"])
			assert.Equal(t, tt.wantReason, event["origin_rejection_reason"])
			assert.Equal(t, tt.wantState, event["origin_header_state"])
			assert.Equal(t, "request-correlation-123", event["request_id"])
			assert.Equal(t, float64(http.StatusForbidden), event["status"])
			assert.Equal(t, "origin_rejected", event["upgrade_error_category"])
			assert.Equal(t, "GET", event["method"])
			assert.Equal(t, "service.example:8080", event["host"])
			assert.Equal(t, "/wss/chrome-service/v1/ws", event["path"])
		})
	}
}

func TestHandleWsConnectionDoesNotMisclassifyOtherUpgradeFailures(t *testing.T) {
	entry, output := newTestRequestLogger("request-correlation-456")
	r := httptest.NewRequest(http.MethodGet, "/ws?query-secret=never-log", nil)
	r.Header.Set("Origin", "https://console.redhat.com")
	r = r.WithContext(context.WithValue(r.Context(), util.IDENTITY_CTX_KEY, &identity.XRHID{Identity: identity.Identity{User: &identity.User{}}}))
	r = r.WithContext(requestlogger.WithLogger(context.WithValue(r.Context(), middleware.RequestIDKey, "request-correlation-456"), entry))
	w := httptest.NewRecorder()

	HandleWsConnection(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, output.String(), "upgrade_error_category")
	assert.Contains(t, output.String(), "upgrade_failed")
	assert.NotContains(t, output.String(), "websocket_origin_rejected")
	assert.NotContains(t, output.String(), "query-secret")
}

func TestRejectedOriginDiagnosticsBoundAndRedactFields(t *testing.T) {
	entry, output := newTestRequestLogger("request-correlation-bounds")
	r := websocketUpgradeRequest()
	r.Header.Set("Origin", "https://user:origin-secret@evil.example/path?token=query-secret#fragment-secret")
	r.Header.Set("User-Agent", strings.Repeat("x", 100)+"\n\ttoken=ua-secret"+strings.Repeat("y", 400))
	r.Header.Set("Cookie", "cookie-secret")
	r.Header.Set("Authorization", "Bearer auth-secret")
	r.Host = "192.0.2.10:8443"
	r.URL.Path = "/ws/token=path-secret"
	r = r.WithContext(context.WithValue(r.Context(), util.IDENTITY_CTX_KEY, &identity.XRHID{Identity: identity.Identity{User: &identity.User{}}}))
	r = r.WithContext(requestlogger.WithLogger(context.WithValue(r.Context(), middleware.RequestIDKey, "request-correlation-bounds"), entry))
	w := httptest.NewRecorder()

	HandleWsConnection(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var event map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(output.String()), &event))
	userAgent, ok := event["user_agent"].(string)
	assert.True(t, ok)
	assert.LessOrEqual(t, len(userAgent), maxDiagnosticFieldLength)
	assert.NotContains(t, userAgent, "\n")
	assert.Contains(t, userAgent, `\n`)
	assert.Equal(t, "<redacted-ip>", event["host"])
	for _, secret := range []string{"origin-secret", "query-secret", "fragment-secret", "ua-secret", "cookie-secret", "auth-secret", "path-secret"} {
		assert.NotContains(t, output.String(), secret)
	}
	assert.NotContains(t, output.String(), "Cookie")
	assert.NotContains(t, output.String(), "Authorization")
	assert.NotContains(t, output.String(), "192.0.2.10")
}

func websocketUpgradeRequest() *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/wss/chrome-service/v1/ws", nil)
	r.Host = "service.example:8080"
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Sec-WebSocket-Version", "13")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	return r
}

func newTestRequestLogger(requestID string) (*logrus.Entry, *strings.Builder) {
	output := &strings.Builder{}
	testLogger := logrus.New()
	testLogger.SetOutput(output)
	testLogger.SetFormatter(&logrus.JSONFormatter{})
	return logrus.NewEntry(testLogger).WithField("request_id", requestID), output
}

func TestHandleWsConnectionRejectsMissingOrInvalidIdentity(t *testing.T) {
	tests := []struct {
		name string
		ctx  func() context.Context
	}{
		{
			name: "no identity in context",
			ctx: func() context.Context {
				return context.Background()
			},
		},
		{
			name: "nil identity pointer in context",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), util.IDENTITY_CTX_KEY, (*identity.XRHID)(nil))
			},
		},
		{
			name: "identity with nil user",
			ctx: func() context.Context {
				xrhid := &identity.XRHID{Identity: identity.Identity{User: nil}}
				return context.WithValue(context.Background(), util.IDENTITY_CTX_KEY, xrhid)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/ws", nil)
			r = r.WithContext(tt.ctx())
			w := httptest.NewRecorder()

			HandleWsConnection(w, r)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
