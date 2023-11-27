package logger

import (
	"bytes"
	"fmt"
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
	request  *http.Request
	buf      *bytes.Buffer
	useColor bool
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &LogEntry{
		StructuredLogger: l,
		request:          r,
		buf:              &bytes.Buffer{},
		useColor:         false,
	}

	reqID := middleware.GetReqID(r.Context())
	if reqID != "" {
		fmt.Fprintf(entry.buf, "[%s] ", reqID)
	}

	fmt.Fprintf(entry.buf, "\"")
	fmt.Fprintf(entry.buf, "%s ", r.Method)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fmt.Fprintf(entry.buf, "%s://%s%s %s\" ", scheme, r.Host, r.RequestURI, r.Proto)

	entry.buf.WriteString("from ")
	entry.buf.WriteString(r.RemoteAddr)
	entry.buf.WriteString(" - ")

	return entry
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	// Do nothing if status code is 200/201/eg and the log level is above Warn (3)
	if (l.LogLevel <= logrus.WarnLevel) && (status < 400) {
		return
	}

	fmt.Fprintf(l.buf, "%03d", status)
	fmt.Fprintf(l.buf, " %dB", bytes)

	l.buf.WriteString(" in ")
	if elapsed < 500*time.Millisecond {
		fmt.Fprintf(l.buf, "%s", elapsed)
	} else if elapsed < 5*time.Second {
		fmt.Fprintf(l.buf, "%s", elapsed)
	} else {
		fmt.Fprintf(l.buf, "%s", elapsed)
	}

	l.Logger.Print(l.buf.String())
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
