package routes

import (
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/RedHatInsights/chrome-service-backend/rest/securitylog"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"cloudevents.json"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MakeWsRoute(sub chi.Router) {
	sub.Get("/", HandleWsConnection)
}

func HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	jwtCookie, err := r.Cookie("cs_jwt")
	if err != nil {
		logrus.Errorln("Unable to find cs_jwt cookie", err)
		securitylog.LogWithReason(r.Context(), "AUTHENTICATE", "websocket", r.RemoteAddr, "failure", "missing JWT cookie")
		return
	}
	identity, err := util.ParseJWTToken(jwtCookie.Value)
	if err != nil {
		logrus.Errorln("Unable to parse jwt token", err)
		securitylog.LogWithReason(r.Context(), "AUTHENTICATE", "websocket", r.RemoteAddr, "failure", "invalid JWT token")
		return
	}

	// REMOVE this, trying to figure out why the connection is being closed
	r.Header.Add("Connection", "upgrade")

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorln("Unable to upgrade WS connection", err)
		return
	}

	client := connectionhub.Client{
		User:         identity.UserId,
		Organization: identity.OrgId,
		Username:     identity.Username,
		Roles:        []string{},
		Conn:         &connectionhub.Connection{Send: make(chan []byte, 256), Ws: ws},
	}
	logrus.Infoln("New client added to the connection hub: ", client.User)
	connectionhub.ConnectionHub.Register <- client
	go client.WritePump()
	client.ReadPump()
}
