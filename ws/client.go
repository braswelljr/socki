package ws

import (
	"bytes"
	"time"

	"github.com/braswelljr/socki/utils"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	logger  = utils.NewLogger()
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	connection *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// NewClient creates a new client
func NewClient(hub *Hub, connection *websocket.Conn) *Client {
	return &Client{
		hub:        hub,
		connection: connection,
		send:       make(chan []byte, maxMessageSize),
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (client *Client) ReadPump() {
	// close connection when function returns
	defer func() {
		// unregister client from hub
		client.hub.unregister <- client

		// close connection
		client.connection.Close()
	}()

	// set read limit, deadline and pong handler
	client.connection.SetReadLimit(maxMessageSize)
	client.connection.SetReadDeadline(time.Now().Add(pongWait))
	client.connection.SetPongHandler(func(string) error {
		client.connection.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// loop through messages
	for {
		// read message from connection
		_, message, err := client.connection.ReadMessage()

		// check for errors
		if err != nil {
			// check for websocket close error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error().Err(err).Msg("Unexpected close error")
			}
			break
		}

		// trim message
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// broadcast message to hub
		client.hub.broadcast <- message
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (client *Client) WritePump() {
	// close connection when function returns
	defer func() {
		// close connection
		client.connection.Close()
	}()

	// loop through messages
	for range client.send {
		// set write deadline
		client.connection.SetWriteDeadline(time.Now().Add(writeWait))

		// check for ok
		if message, ok := <-client.send; !ok {
			// write close message
			client.connection.WriteMessage(websocket.CloseMessage, []byte{})
			return
		} else {
			// create websocket message
			writer, err := client.connection.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error().Err(err).Msg("Error creating websocket message")
				return
			}

			// write message
			writer.Write(message)

			// loop through messages
			n := len(client.send)

			for i := 0; i < n; i++ {
				// write message
				writer.Write(newline)
				writer.Write(<-client.send)
			}

			// close writer
			if err := writer.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing writer")
				return
			}
		}

	}
}
