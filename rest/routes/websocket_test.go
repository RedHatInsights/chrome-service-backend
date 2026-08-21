package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
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
		{"subdomain of console.redhat.com", "https://beta.console.redhat.com", true},
		{"subdomain of console.stage.redhat.com", "https://beta.console.stage.redhat.com", true},
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
