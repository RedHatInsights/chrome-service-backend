package logger

import (
	"context"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

func LogFor(ctx context.Context) *logrus.Entry {
	if entry, ok := ctx.Value(loggerKey).(*logrus.Entry); ok {
		return entry
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func WithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, entry)
}

func InjectLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := logrus.WithField("request_id", middleware.GetReqID(r.Context()))
		ctx := WithLogger(r.Context(), entry)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func EnrichLoggerWithIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		entry := LogFor(ctx)

		if id, ok := ctx.Value(util.IDENTITY_CTX_KEY).(*identity.XRHID); ok && id != nil {
			entry = entry.WithFields(logrus.Fields{
				"org_id":  id.Identity.OrgID,
				"account": id.Identity.AccountNumber,
			})
			ctx = WithLogger(ctx, entry)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
