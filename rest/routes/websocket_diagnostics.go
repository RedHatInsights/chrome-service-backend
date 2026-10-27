package routes

import (
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

const maxDiagnosticFieldLength = 256

var (
	sensitiveAssignment = regexp.MustCompile(`(?i)(authorization|token|secret|password|credential|cookie|identity|api[_-]?key|access[_-]?key|private[_-]?key|session|jwt)(\s*[:=]\s*)[^\s,;&]+`)
	bearerCredential    = regexp.MustCompile(`(?i)bearer\s+[^\s,;]+`)
	jwtCredential       = regexp.MustCompile(`\b[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\b`)
	ipv4Address         = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	ipv6Address         = regexp.MustCompile(`(?i)(?:[0-9a-f]{0,4}:){2,7}[0-9a-f]{0,4}`)
)

type originHeaderInfo struct {
	state           string
	representation  string
	rejectionReason string
}

// originRejectionReason is the single decision used by CheckOrigin and its
// diagnostic path. Empty string means the origin is accepted.
func originRejectionReason(r *http.Request) string {
	return inspectOriginHeader(r).rejectionReason
}

func inspectOriginHeader(r *http.Request) originHeaderInfo {
	info := originHeaderInfo{}
	value := r.Header.Get("Origin")
	var present bool
	var nullValue bool
	for name, values := range r.Header {
		if strings.EqualFold(name, "Origin") {
			present = true
			if values == nil {
				nullValue = true
			}
			break
		}
	}

	switch {
	case !present:
		info.state = "missing"
		info.rejectionReason = "missing"
		info.representation = "<missing>"
		return info
	case nullValue:
		info.state = "null"
		info.rejectionReason = "missing"
		info.representation = "<null>"
		return info
	case value == "":
		info.state = "empty"
		info.rejectionReason = "missing"
		info.representation = "<empty>"
		return info
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		info.state = "present"
		info.rejectionReason = "malformed"
		info.representation = "<malformed>"
		return info
	}

	info.state = "present"
	hostKind := "<unlisted-host>"
	if isAllowedOriginHost(parsed.Hostname()) {
		hostKind = "<allowed-host>"
	}
	info.representation = safeDiagnosticString(strings.ToLower(parsed.Scheme)+"://"+hostKind, 64)
	if parsed.Scheme != "https" {
		info.rejectionReason = "disallowed_scheme"
	} else if !isAllowedOriginHost(parsed.Hostname()) {
		info.rejectionReason = "disallowed_host"
	}
	return info
}

func isAllowedOriginHost(host string) bool {
	for _, suffix := range allowedOriginSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return host == "console.redhat.com" || host == "console.stage.redhat.com" || host == "console.dev.redhat.com"
}

func logRejectedWebSocketOrigin(r *http.Request, reason string) {
	origin := inspectOriginHeader(r)
	requestlogger := logger.LogFor(r.Context())
	requestlogger.WithFields(logrus.Fields{
		"event":                   "websocket_origin_rejected",
		"request_id":              safeDiagnosticString(middleware.GetReqID(r.Context()), 128),
		"method":                  safeDiagnosticString(r.Method, 32),
		"host":                    safeRequestHost(r.Host),
		"path":                    safeDiagnosticString(r.URL.Path, maxDiagnosticFieldLength),
		"user_agent":              safeDiagnosticString(r.UserAgent(), maxDiagnosticFieldLength),
		"origin_header_state":     origin.state,
		"origin_representation":   origin.representation,
		"origin_rejection_reason": reason,
		"status":                  http.StatusForbidden,
		"upgrade_error_category":  "origin_rejected",
	}).Warn("WebSocket upgrade rejected")
}

func safeRequestHost(host string) string {
	hostname := host
	port := ""
	if parsedHost, parsedPort, err := net.SplitHostPort(host); err == nil {
		hostname = parsedHost
		if value, err := strconv.Atoi(parsedPort); err == nil && value >= 0 && value <= 65535 {
			port = parsedPort
		} else {
			return "<redacted-host>"
		}
	} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		hostname = strings.Trim(host, "[]")
	} else if strings.Contains(host, ":") {
		return "<redacted-host>"
	}
	if net.ParseIP(hostname) != nil {
		return "<redacted-ip>"
	}
	for _, char := range hostname {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '.' || char == '-') {
			return "<redacted-host>"
		}
	}
	if port != "" {
		hostname += ":" + port
	}
	return safeDiagnosticString(hostname, 255)
}

// safeDiagnosticString redacts common credential forms, escapes control
// characters, and bounds the resulting log field.
func safeDiagnosticString(value string, maxLength int) string {
	value = sensitiveAssignment.ReplaceAllString(value, "$1$2<redacted>")
	value = bearerCredential.ReplaceAllString(value, "Bearer <redacted>")
	value = jwtCredential.ReplaceAllString(value, "<redacted-jwt>")
	value = ipv4Address.ReplaceAllStringFunc(value, func(candidate string) string {
		if net.ParseIP(candidate) != nil {
			return "<redacted-ip>"
		}
		return candidate
	})
	value = ipv6Address.ReplaceAllStringFunc(value, func(candidate string) string {
		if net.ParseIP(candidate) != nil {
			return "<redacted-ip>"
		}
		return candidate
	})
	escaped := strings.Trim(strconv.QuoteToASCII(value), `"`)
	if len(escaped) > maxLength {
		escaped = escaped[:maxLength]
	}
	return escaped
}
