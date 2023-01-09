package connectionhub

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type clients = map[string]*Client

type Client struct {
	User         string
	Organization string
	Roles        []string
	Conn         *Connection
}

type MessageDestinations struct {
	Users         []string
	Roles         []string
	Organizations []string
}

type Message struct {
	Broadcast    bool
	Data         []byte
	Destinations MessageDestinations
	Origin       string
}

type WsMessage struct {
	Broadcast     bool                   `json:"broadcast"`
	Users         []string               `json:"users"`
	Roles         []string               `json:"roles"`
	Organizations []string               `json:"organizations"`
	Payload       map[string]interface{} `json:"payload"`
}

type ConnectionNamespaces struct {
	// index rooms clients by connections to allow better access
	Roles        map[string]map[*Connection]*Client
	Organization map[string]map[*Connection]*Client
}

type connectionHub struct {
	Rooms      ConnectionNamespaces
	Emit       chan Message
	Broadcast  chan Message
	Register   chan Client
	Unregister chan Client
	Clients    clients
}

var ConnectionHub = connectionHub{
	Rooms: ConnectionNamespaces{
		Roles:        make(map[string]map[*Connection]*Client),
		Organization: make(map[string]map[*Connection]*Client),
	},
	Emit:       make(chan Message),
	Broadcast:  make(chan Message),
	Register:   make(chan Client),
	Unregister: make(chan Client),
	Clients:    make(clients),
}

func registerClientRoles(c Client, h *connectionHub) {
	for _, role := range c.Roles {
		if h.Rooms.Roles[role] == nil {
			h.Rooms.Roles[role] = make(map[*Connection]*Client)
		}
		h.Rooms.Roles[role][c.Conn] = &c
	}
}

func registerClientOrg(c Client, h *connectionHub) {
	if h.Rooms.Organization[c.Organization] == nil {
		h.Rooms.Organization[c.Organization] = make(map[*Connection]*Client)
	}
	h.Rooms.Organization[c.Organization][c.Conn] = &c
}

func registerClient(c Client, h *connectionHub) {
	if h.Clients[c.User] == nil {
		h.Clients[c.User] = &c
	}
	registerClientRoles(c, h)
	registerClientOrg(c, h)
	logrus.Debugln("new client connected", c)
}

func unregisterClientOrg(c Client, h *connectionHub) {
	if h.Rooms.Organization[c.Organization] != nil {
		delete(h.Rooms.Organization[c.Organization], c.Conn)
	}
}

func unregisterClientRoles(c Client, h *connectionHub) {
	for _, role := range c.Roles {
		if h.Rooms.Roles[role] != nil {
			delete(h.Rooms.Roles[role], c.Conn)
		}
	}
}

func unregisterClient(c Client, h *connectionHub) {
	unregisterClientRoles(c, h)
	unregisterClientOrg(c, h)
	if h.Clients[c.User] != nil {
		delete(h.Clients, c.User)
	}
}

func broadcast(m Message, h *connectionHub) {
	for _, c := range h.Clients {
		select {
		case c.Conn.Send <- m.Data:
		default:
			// if message fails to be sent, remove the client as it is no longer active
			close(c.Conn.Send)
			unregisterClient(*c, h)
		}
	}
}

func emitMessage(m Message, h *connectionHub) {
	connections := make(map[*Connection]*Client)
	// get all individual connections
	for _, cid := range m.Destinations.Users {
		connections[h.Clients[cid].Conn] = h.Clients[cid]
	}

	// get all connections from rooms
	for _, rid := range m.Destinations.Roles {
		for conn, c := range h.Rooms.Roles[rid] {
			connections[conn] = c
		}
	}

	// get all connections from organizations
	for _, oid := range m.Destinations.Organizations {
		for conn, c := range h.Rooms.Organization[oid] {
			connections[conn] = c
		}
	}

	// distribute message to connection channels
	for conn, client := range connections {
		select {
		case conn.Send <- m.Data:
		default:
			unregisterClient(*client, h)
		}
	}

}

func (h *connectionHub) Run() {
	for {
		select {
		case c := <-h.Register:
			registerClient(c, h)
		case c := <-h.Unregister:
			unregisterClient(c, h)
		case m := <-h.Broadcast:
			broadcast(m, h)
		case m := <-h.Emit:
			emitMessage(m, h)
		}
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func (c Client) ReadPump() {
	conn := c.Conn
	// close connection after client is was removed
	defer func() {
		ConnectionHub.Unregister <- c
		conn.Ws.Close()
	}()
	// configure ws connection
	conn.Ws.SetReadLimit(maxMessageSize)
	conn.Ws.SetReadDeadline(time.Now().Add(pongWait))
	conn.Ws.SetPongHandler(func(appData string) error {
		conn.Ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := conn.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				logrus.Errorln("error: %v", err)
			}
			logrus.Errorln("Unable to read incomming WS message: ", err)
			break
		}
		var messagePayload WsMessage
		json.Unmarshal(msg, &messagePayload)
		if err != nil {
			logrus.Errorln("Unable to unmarshall incomming WS message: ", err)
			break
		}

		var message Message
		message.Data = msg
		if messagePayload.Broadcast {
			message.Broadcast = true
			ConnectionHub.Broadcast <- message
		} else {
			message.Destinations = MessageDestinations{
				Users:         messagePayload.Users,
				Roles:         messagePayload.Roles,
				Organizations: messagePayload.Organizations,
			}
			ConnectionHub.Emit <- message
		}
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}

// pump incomming messages to connection hub
func (c Client) WritePump() {
	conn := c.Conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Ws.Close()
	}()

	for {
		select {
		case message, ok := <-conn.Send:
			if !ok {
				logrus.Errorln("sending message has failed", message)
				conn.write(websocket.CloseMessage, []byte{})
			}
			if err := conn.write(websocket.TextMessage, message); err != nil {
				logrus.Errorln("Unable to write message to WS connection: ", err)
				return
			}
		case <-ticker.C:
			if err := conn.write(websocket.PingMessage, []byte{}); err != nil {
				logrus.Errorln("Heart beat frame failed to be send: ", err)
				return
			}
		}
	}
}
