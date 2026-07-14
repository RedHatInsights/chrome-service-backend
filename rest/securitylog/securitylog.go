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

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

// Log emits a structured security event.
// SEC-MON-REQ-1 compliance (EOI-1 pii_manipulation, EOI-2 system_object_manipulation,
// EOI-3 admin_action)
func Log(ctx context.Context, action, resourceType, resourceID, outcome string) {
	fields := logrus.Fields{
		"event":         "security",
		"action":        action,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"outcome":       outcome,
	}

	addPrincipal(ctx, fields)

	if outcome == "failure" {
		logrus.WithFields(fields).Warn("security_event")
	} else {
		logrus.WithFields(fields).Info("security_event")
	}
}

// LogWithReason emits a security event with an additional reason field.
// SEC-MON-REQ-1 compliance (EOI-7 invalid_login, EOI-8 authorization_failure,
// EOI-11 warnings_or_errors)
func LogWithReason(ctx context.Context, action, resourceType, resourceID, outcome, reason string) {
	fields := logrus.Fields{
		"event":         "security",
		"action":        action,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"outcome":       outcome,
		"reason":        reason,
	}

	addPrincipal(ctx, fields)

	if outcome == "failure" {
		logrus.WithFields(fields).Warn("security_event")
	} else {
		logrus.WithFields(fields).Info("security_event")
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
	logrus.WithFields(logrus.Fields{
		"event":         "security",
		"action":        "SHUTDOWN",
		"resource_type": "process",
		"resource_id":   serviceName,
		"outcome":       "failure",
		"reason":        reason,
	}).Error("security_event")
}

// addPrincipal extracts user_id and org_id from the request context and adds
// them to the log fields. When identity is not available (e.g. unauthenticated
// requests), the principal fields are omitted.
func addPrincipal(ctx context.Context, fields logrus.Fields) {
	if ctx == nil {
		return
	}
	if id, ok := ctx.Value(util.IDENTITY_CTX_KEY).(*identity.XRHID); ok && id != nil {
		fields["user_id"] = id.Identity.User.UserID
		fields["org_id"] = id.Identity.OrgID
	}
}
