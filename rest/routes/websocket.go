package routes

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
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
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "https" {
		return false
	}
	host := parsed.Hostname()
	for _, suffix := range allowedOriginSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return host == "console.redhat.com" || host == "console.stage.redhat.com"
}

func MakeWsRoute(sub chi.Router) {
	sub.Get("/", HandleWsConnection)
}

func HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	xrhid, ok := r.Context().Value(util.IDENTITY_CTX_KEY).(*identity.XRHID)
	if !ok || xrhid == nil || xrhid.Identity.User == nil {
		logrus.Errorln("WebSocket connection rejected: missing or invalid X-RH-Identity")
		securitylog.LogWithReason(r.Context(), "AUTHENTICATE", "websocket", r.RemoteAddr, "failure", "missing verified identity")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorln("Unable to upgrade WS connection", err)
		return
	}

	client := connectionhub.Client{
		User:         xrhid.Identity.User.UserID,
		Organization: xrhid.Identity.OrgID,
		Username:     xrhid.Identity.User.Username,
		Roles:        []string{},
		Conn:         &connectionhub.Connection{Send: make(chan []byte, 256), Ws: ws},
	}
	logrus.Infoln("New client added to the connection hub: ", client.User)
	connectionhub.ConnectionHub.Register <- client
	go client.WritePump()
	client.ReadPump()
}
