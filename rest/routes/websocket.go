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

func debugHeaders(r *http.Request) {
	fmt.Println("Headers")
	fmt.Println("Request method: ", r.Method)
	for k, v := range r.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
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
	debugHeaders(r)

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

func EmitMessage(w http.ResponseWriter, r *http.Request) {
	var p WSRequestPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	logrus.Infoln("Attempting to emit new broadcast message", p)
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
	event := cloudevents.WrapPayload(p.Payload, cloudevents.URI(r.Host+r.URL.Path), p.Id, p.Type)
	dctErr := event.DataContentType.IsValid()
	if dctErr != nil {
		logrus.Errorln(dctErr)
		payload := make(map[string]string)
		payload["msg"] = "The Data Content Type needs to be in application/json format!"
		response, _ := json.Marshal(payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}
	svErr := event.SpecVersion.IsValid()
	if svErr != nil {
		logrus.Errorln(svErr)
		payload := make(map[string]string)
		payload["msg"] = "Spec version needs to be 1.0.2!"
		response, _ := json.Marshal(payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}
	uriErr := event.Source.IsValid()
	if uriErr != nil {
		logrus.Errorln(svErr)
		payload := make(map[string]string)
		payload["msg"] = "Invalid URI!"
		response, _ := json.Marshal(payload)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}
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
