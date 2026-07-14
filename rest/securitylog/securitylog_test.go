package securitylog

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// captureOutput captures logrus output during fn execution.
func captureOutput(fn func()) string {
	origOutput := logrus.StandardLogger().Out
	origLevel := logrus.GetLevel()
	origFormatter := logrus.StandardLogger().Formatter

	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	defer func() {
		logrus.SetOutput(origOutput)
		logrus.SetLevel(origLevel)
		logrus.SetFormatter(origFormatter)
	}()
	fn()
	return buf.String()
}

func TestLogSuccess(t *testing.T) {
	output := captureOutput(func() {
		Log(context.Background(), "CREATE", "dashboard_template", "42", "success")
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "CREATE")
	assert.Contains(t, output, "dashboard_template")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "success")
	assert.Contains(t, output, "level=info")
}

func TestLogFailure(t *testing.T) {
	output := captureOutput(func() {
		Log(context.Background(), "DELETE", "favorite_page", "/foo", "failure")
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "DELETE")
	assert.Contains(t, output, "level=warning")
}

func TestLogWithReason(t *testing.T) {
	output := captureOutput(func() {
		LogWithReason(context.Background(), "AUTHENTICATE", "api_request", "/api/test", "failure", "missing identity header")
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "AUTHENTICATE")
	assert.Contains(t, output, "missing identity header")
	assert.Contains(t, output, "level=warning")
}

func TestLogWithReasonSuccess(t *testing.T) {
	output := captureOutput(func() {
		LogWithReason(context.Background(), "UPDATE", "user_identity", "user1", "success", "preview updated")
	})

	assert.Contains(t, output, "level=info")
}

func TestLogStartup(t *testing.T) {
	output := captureOutput(func() {
		LogStartup("chrome-service-backend", 8000)
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "STARTUP")
	assert.Contains(t, output, "process")
	assert.Contains(t, output, "chrome-service-backend")
	assert.Contains(t, output, "level=info")
}

func TestLogShutdown(t *testing.T) {
	output := captureOutput(func() {
		LogShutdown("chrome-service-backend", "server stopped")
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "SHUTDOWN")
	assert.Contains(t, output, "server stopped")
	assert.Contains(t, output, "level=error")
}

func TestLogWithPrincipal(t *testing.T) {
	ctx := context.WithValue(context.Background(), util.IDENTITY_CTX_KEY, &identity.XRHID{
		Identity: identity.Identity{
			OrgID: "org-123",
			User:  &identity.User{UserID: "user-456"},
		},
	})

	output := captureOutput(func() {
		Log(ctx, "UPDATE", "self_report", "user-456", "success")
	})

	assert.Contains(t, output, "user-456")
	assert.Contains(t, output, "org-123")
}

func TestLogServiceAccountIdentity(t *testing.T) {
	// Service accounts have nil User — must not panic
	ctx := context.WithValue(context.Background(), util.IDENTITY_CTX_KEY, &identity.XRHID{
		Identity: identity.Identity{
			OrgID: "org-789",
			User:  nil,
		},
	})

	output := captureOutput(func() {
		Log(ctx, "CREATE", "dashboard_template", "1", "success")
	})

	assert.Contains(t, output, "security_event")
	assert.Contains(t, output, "org-789")
	assert.False(t, strings.Contains(output, "user_id"))
}

func TestLogWithoutReasonOmitsField(t *testing.T) {
	output := captureOutput(func() {
		Log(context.Background(), "CREATE", "test", "1", "success")
	})

	assert.Contains(t, output, "security_event")
	// Log() calls LogWithReason with empty reason — reason field should be omitted
	assert.False(t, strings.Contains(output, "reason"))
}

func TestLogNilContext(t *testing.T) {
	// Should not panic with nil context
	output := captureOutput(func() {
		Log(nil, "READ", "test", "1", "success")
	})

	assert.Contains(t, output, "security_event")
	// No principal fields expected
	assert.False(t, strings.Contains(output, "user_id"))
}

func TestLogNoIdentityInContext(t *testing.T) {
	output := captureOutput(func() {
		Log(context.Background(), "CREATE", "test", "1", "success")
	})

	assert.Contains(t, output, "security_event")
	// No principal fields when identity not in context
	assert.False(t, strings.Contains(output, "user_id"))
}
