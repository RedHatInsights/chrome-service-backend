// Package securitylog provides structured security event logging per
// SEC-MON-REQ-1 compliance (Events of Interest).
//
// Every security log includes the five required fields:
//   - action:        CREATE, READ, UPDATE, DELETE, or a custom action
//   - resource_type: type of object being operated on
//   - resource_id:   identifier of the specific object
//   - outcome:       "success" or "failure"
//   - principal:     user_id and org_id extracted from request context
package securitylog

import (
	"context"
	"unicode"

	requestlogger "github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

// Log emits a structured security event.
// SEC-MON-REQ-1 compliance (EOI-1 pii_manipulation, EOI-2 system_object_manipulation,
// EOI-3 admin_action)
func Log(ctx context.Context, action, resourceType, resourceID, outcome string) {
	LogWithReason(ctx, action, resourceType, resourceID, outcome, "")
}

// LogWithReason emits a security event with an additional reason field.
// SEC-MON-REQ-1 compliance (EOI-7 invalid_login, EOI-8 authorization_failure,
// EOI-11 warnings_or_errors)
func LogWithReason(ctx context.Context, action, resourceType, resourceID, outcome, reason string) {
	fields := logrus.Fields{
		"event":         "security",
		"action":        boundedField(action, 32),
		"resource_type": boundedField(resourceType, 64),
		"resource_id":   boundedField(resourceID, 256),
		"outcome":       boundedField(outcome, 16),
	}

	if reason != "" {
		fields["reason"] = boundedField(reason, 128)
	}

	addPrincipal(ctx, fields)

	if outcome == "failure" {
		requestlogger.LogFor(ctx).WithFields(fields).Warn("security_event")
	} else {
		requestlogger.LogFor(ctx).WithFields(fields).Info("security_event")
	}
}

// LogStartup emits a process startup event.
// SEC-MON-REQ-1 compliance (EOI-5 process_status)
func LogStartup(serviceName string, port int) {
	logrus.WithFields(logrus.Fields{
		"event":         "security",
		"action":        "STARTUP",
		"resource_type": "process",
		"resource_id":   serviceName,
		"outcome":       "success",
		"port":          port,
	}).Info("security_event")
}

// LogShutdown emits a process shutdown event.
// SEC-MON-REQ-1 compliance (EOI-5 process_status)
func LogShutdown(serviceName, reason string) {
	LogShutdownWithOutcome(serviceName, reason, false)
}

// LogShutdownWithOutcome marks unexpected termination or a failed drain as failure.
func LogShutdownWithOutcome(serviceName, reason string, failed bool) {
	outcome := "success"
	level := logrus.InfoLevel
	if failed {
		outcome = "failure"
		level = logrus.ErrorLevel
	}
	logrus.WithFields(logrus.Fields{
		"event":         "security",
		"action":        "SHUTDOWN",
		"resource_type": "process",
		"resource_id":   serviceName,
		"outcome":       outcome,
		"reason":        boundedField(reason, 128),
	}).Log(level, "security_event")
}

// addPrincipal extracts user_id and org_id from the request context and adds
// them to the log fields. When identity is not available (e.g. unauthenticated
// requests), the principal fields are omitted.
func addPrincipal(ctx context.Context, fields logrus.Fields) {
	if ctx == nil {
		return
	}
	if id, ok := ctx.Value(util.IDENTITY_CTX_KEY).(*identity.XRHID); ok && id != nil {
		if id.Identity.User != nil {
			fields["user_id"] = boundedField(id.Identity.User.UserID, 128)
		}
		fields["org_id"] = boundedField(id.Identity.OrgID, 128)
	}
}

func boundedField(value string, maxLength int) string {
	var safe []rune
	for _, char := range value {
		if len(safe) >= maxLength {
			break
		}
		if unicode.IsControl(char) {
			char = '_'
		}
		safe = append(safe, char)
	}
	return string(safe)
}
