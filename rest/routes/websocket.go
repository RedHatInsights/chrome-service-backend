package routes

import (
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/securitylog"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
)

var allowedOriginSuffixes = []string{
	".console.redhat.com",
	".console.stage.redhat.com",
	".console.dev.redhat.com",
	".foo.redhat.com",
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"cloudevents.json"},
	CheckOrigin:     checkOrigin,
}

// checkOrigin validates the WebSocket upgrade Origin header.
// Must stay in sync with CORS AllowedOrigins in main.go.
func checkOrigin(r *http.Request) bool {
	return originRejectionReason(r) == ""
}

func MakeWsRoute(sub chi.Router) {
	sub.Get("/", HandleWsConnection)
}

func HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	xrhid, ok := r.Context().Value(util.IDENTITY_CTX_KEY).(*identity.XRHID)
	if !ok || xrhid == nil || xrhid.Identity.User == nil {
		securitylog.LogWithReason(r.Context(), "AUTHENTICATE", "websocket", "websocket", "failure", "missing verified identity")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Copy the upgrader so the request-local callback can record why this
	// specific upgrade was rejected without sharing mutable state across requests.
	requestUpgrader := upgrader
	originReason := ""
	requestUpgrader.CheckOrigin = func(req *http.Request) bool {
		originReason = originRejectionReason(req)
		return originReason == ""
	}
	ws, err := requestUpgrader.Upgrade(w, r, nil)
	if err != nil {
		if originReason != "" {
			logRejectedWebSocketOrigin(r, originReason)
		} else {
			logger.LogFor(r.Context()).WithFields(logrus.Fields{
				"upgrade_error_category": "upgrade_failed",
			}).WithError(err).Error("Unable to upgrade WS connection")
		}
		return
	}

	client := connectionhub.Client{
		User:         xrhid.Identity.User.UserID,
		Organization: xrhid.Identity.OrgID,
		Username:     xrhid.Identity.User.Username,
		Roles:        []string{},
		Conn:         &connectionhub.Connection{Send: make(chan []byte, 256), Ws: ws},
	}
	logger.LogFor(r.Context()).Debug("websocket client registration requested")
	connectionhub.ConnectionHub.Register <- client
	logger.LogFor(r.Context()).Debug("websocket client registration accepted")
	go client.WritePump()
	client.ReadPump()
}
