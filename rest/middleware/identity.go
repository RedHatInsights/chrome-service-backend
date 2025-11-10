package middleware

import (
	"context"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
)

func InjectUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := r.Context().Value(util.IDENTITY_CTX_KEY).(*identity.XRHID)
		userId := identity.Identity.User.UserID
		skipCache := false
		p := r.URL.Query().Get("skip-identity-cache")
		if p == "true" {
			skipCache = true
		}
		userIdentity, err := service.CreateIdentity(userId, skipCache)
		if err != nil {
			panic(err)
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, util.USER_CTX_KEY, userIdentity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
