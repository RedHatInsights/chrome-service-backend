package logger

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

const maxCorrelationValueLength = 128

func safeCorrelationValue(value string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	var safe strings.Builder
	for _, char := range value {
		if safe.Len()+utf8.RuneLen(char) > maxLength {
			break
		}
		if unicode.IsControl(char) {
			char = '_'
		}
		safe.WriteRune(char)
	}
	return safe.String()
}

type requestLogger struct {
	mu     sync.RWMutex
	fields logrus.Fields
}

func (l *requestLogger) addFields(fields logrus.Fields) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, value := range fields {
		l.fields[key] = value
	}
}

func (l *requestLogger) entry() *logrus.Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	fields := make(logrus.Fields, len(l.fields))
	for key, value := range l.fields {
		fields[key] = value
	}
	return logrus.WithFields(fields)
}

func LogFor(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	if entry, ok := ctx.Value(loggerKey).(*requestLogger); ok && entry != nil {
		return entry.entry()
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func WithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	fields := logrus.Fields{}
	if entry != nil {
		for key, value := range entry.Data {
			fields[key] = value
		}
	}
	return context.WithValue(ctx, loggerKey, &requestLogger{fields: fields})
}

func AddFields(ctx context.Context, fields logrus.Fields) {
	if ctx == nil {
		return
	}
	if request, ok := ctx.Value(loggerKey).(*requestLogger); ok && request != nil {
		request.addFields(fields)
	}
}

func InjectLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := logrus.WithField("request_id", safeCorrelationValue(middleware.GetReqID(r.Context()), maxCorrelationValueLength))
		ctx := WithLogger(r.Context(), entry)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func EnrichLoggerWithIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if id, ok := ctx.Value(util.IDENTITY_CTX_KEY).(*identity.XRHID); ok && id != nil {
			fields := logrus.Fields{
				"org_id":  safeCorrelationValue(id.Identity.OrgID, maxCorrelationValueLength),
				"account": safeCorrelationValue(id.Identity.AccountNumber, maxCorrelationValueLength),
			}
			if id.Identity.User != nil {
				fields["user_id"] = safeCorrelationValue(id.Identity.User.UserID, maxCorrelationValueLength)
			}
			AddFields(ctx, fields)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
