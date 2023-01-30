package routes

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MakeWsRoute(sub chi.Router) {
	sub.Get("/", HandleWsConnection)
}

func HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	clientId := fmt.Sprint(rand.Int())
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorln("Unable to upgrade WS connection", err)
		return
	}

	client := connectionhub.Client{
		User:         clientId,
		Organization: "foo",
		Roles:        []string{},
		Conn:         &connectionhub.Connection{Send: make(chan []byte, 256), Ws: ws},
	}
	connectionhub.ConnectionHub.Register <- client
	go client.WritePump()
	client.ReadPump()
}
