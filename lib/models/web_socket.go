package models

import (
	"fmt"

	"golang.org/x/net/websocket"
)

type WebSocketStruct struct {
	Origin string `json:"origin"`
	URL    string `json:"url"`
}

func (ws WebSocketStruct) Dial() (*websocket.Conn, error) {
	fmt.Println("WebSocket Dial")
	return websocket.Dial(ws.URL, "", ws.Origin)
}
