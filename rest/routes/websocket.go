package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/cloudevents"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WSRequestPayload struct {
	connectionhub.WsMessage
	Type string `json:"type"`
	Id   string `json:"id"`
}

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

func BroadcastMessage(sub chi.Router) {
	sub.Post("/", EmitMessage)
}

func HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	jwtCookie, err := r.Cookie("cs_jwt")
	if err != nil {
		logrus.Errorln("Unable to find cs_jwt cookie", err)
		return
	}
	identity, err := util.ParseJWTToken(jwtCookie.Value)
	if err != nil {
		logrus.Errorln("Unable to parse jwt token", err)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Errorln("Unable to upgrade WS connection", err)
		return
	}

	client := connectionhub.Client{
		User:         identity.UserId,
		Organization: identity.OrgId,
		Roles:        []string{},
		Conn:         &connectionhub.Connection{Send: make(chan []byte, 256), Ws: ws},
	}
	connectionhub.ConnectionHub.Register <- client
	go client.WritePump()
	client.ReadPump()
}

func EmitMessage(w http.ResponseWriter, r *http.Request) {
	var p WSRequestPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		logrus.Errorln(err)
		payload := make(map[string]string)
		payload["msg"] = "Unable to decode payload!"
		response, _ := json.Marshal(payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}
	event := cloudevents.WrapPayload(p.Payload, r.Host+r.URL.Path, p.Id, p.Type)
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Errorln(err)
		payload := make(map[string]string)
		payload["msg"] = "Unable to decode payload!"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("p", p)
	newMessage := connectionhub.Message{
		Destinations: connectionhub.MessageDestinations{
			Users:         p.Users,
			Roles:         p.Roles,
			Organizations: p.Organizations,
		},
		Broadcast: p.Broadcast,
		Data:      data,
	}
	if newMessage.Broadcast {
		connectionhub.ConnectionHub.Broadcast <- newMessage
	} else {
		connectionhub.ConnectionHub.Emit <- newMessage
	}
	w.WriteHeader(http.StatusOK)
}
