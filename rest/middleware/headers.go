package middleware

import (
	"context"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/sirupsen/logrus"
)

func ParseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(util.XRHIDENTITY)
		ctx := r.Context()
		if header == "" {
			errString := "Missing authentication"
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(errString))
			logrus.Errorf("missing the %s header", util.XRHIDENTITY)
			return
		} else {
			identity, err := util.ParseXRHIdentityHeader(header)
			if err != nil {
				logrus.Errorln("Error parsing X-RH-IDENTITY header: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal server error"))
				return
			}
			ctx = context.WithValue(ctx, util.IDENTITY_CTX_KEY, identity)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
