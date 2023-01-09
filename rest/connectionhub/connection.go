package connectionhub

import "github.com/gorilla/websocket"

type Connection struct {
	Ws   *websocket.Conn
	Send chan []byte
}
