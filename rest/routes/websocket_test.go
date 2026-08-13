package routes

import (
	"net/http"
	"testing"

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
