package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// update buffer size for both read and writes to 1024 bytes
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewWebSocketServer creates a new websocket server
func NewWebSocketServer(w http.ResponseWriter, r *http.Request) {
	// connect with the upgrader
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	hub := NewHub()

	// start hub
	go hub.Run()

	// create a new client
	client := NewClient(hub, connection)

	// register client with hub
	client.hub.register <- client

	// start reading from client connection
	go client.ReadPump()
	go client.WritePump()
}
