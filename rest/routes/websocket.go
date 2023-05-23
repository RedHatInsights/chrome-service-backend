package routes

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/rest/cloudEvents"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
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

func BroadcastMessage(sub chi.Router) {
	sub.Post("/", EmitMessage)
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

func EmitMessage(w http.ResponseWriter, r *http.Request) {
	var p connectionhub.WsMessage
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
	event := cloudEvents.WrapPayload(p.Payload, "emit-message-endpoint", "foo-bar")
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
