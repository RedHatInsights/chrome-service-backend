package logger

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

const maxRouteLength = 256

type StructuredLogger struct {
	Logger   middleware.LoggerInterface
	LogLevel logrus.Level
}

type LogEntry struct {
	*StructuredLogger
	request *http.Request
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &LogEntry{
		StructuredLogger: l,
		request:          r,
	}
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	level := completionLevel(status)
	if level > l.LogLevel {
		return
	}

	LogFor(l.request.Context()).WithFields(logrus.Fields{
		"method":     safeCorrelationValue(l.request.Method, 16),
		"route":      requestRoute(l.request),
		"status":     status,
		"bytes":      bytes,
		"elapsed_ms": elapsed.Milliseconds(),
	}).Log(level, "request completed")
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	fields := logrus.Fields{"panic_type": fmt.Sprintf("%T", v)}
	if len(stack) > 8192 {
		stack = stack[:8192]
	}
	if len(stack) > 0 {
		fields["stack"] = string(stack)
	}
	LogFor(l.request.Context()).WithFields(fields).Error("request panic")
}

func completionLevel(status int) logrus.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return logrus.ErrorLevel
	case status >= http.StatusBadRequest:
		return logrus.WarnLevel
	default:
		return logrus.InfoLevel
	}
}

func requestRoute(r *http.Request) string {
	if r == nil {
		return "unmatched"
	}
	pattern := ""
	if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
		pattern = routeContext.RoutePattern()
	}
	if pattern == "" {
		return "unmatched"
	}
	var route strings.Builder
	for _, char := range pattern {
		if route.Len()+utf8.RuneLen(char) > maxRouteLength {
			break
		}
		if unicode.IsControl(char) {
			route.WriteByte('_')
		} else {
			route.WriteRune(char)
		}
	}
	if route.Len() == 0 {
		return "unmatched"
	}
	return route.String()
}

func NewLogger(opts *config.ChromeServiceConfig, logger *logrus.Logger) *StructuredLogger {
	logLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	return &StructuredLogger{
		Logger:   logger,
		LogLevel: logLevel,
	}
}
