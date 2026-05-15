package logger

import (
	"net/http"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

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
	if (l.LogLevel <= logrus.WarnLevel) && (status < 400) {
		return
	}

	LogFor(l.request.Context()).WithFields(logrus.Fields{
		"method":      l.request.Method,
		"path":        l.request.RequestURI,
		"status":      status,
		"bytes":       bytes,
		"elapsed_ms":  elapsed.Milliseconds(),
		"remote_addr": l.request.RemoteAddr,
	}).Info("request completed")
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	middleware.PrintPrettyStack(v)
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
