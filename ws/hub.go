package ws

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub is an instance of the Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run runs and broadcast clients messages on hub
func (hub *Hub) Run() {
	// loop through channels and handle messages
	for {
		// check for multiple valid options on channels
		select {
		case client := <-hub.register:
			// register client
			hub.clients[client] = true

		case client := <-hub.unregister:
			// check if client is registered
			if _, ok := hub.clients[client]; ok {
				// delete client from map
				delete(hub.clients, client)

				// close client send channel
				close(client.send)
			}

		case message := <-hub.broadcast:
			// loop through registered clients
			for client := range hub.clients {
				// send message to client
				select {
				case client.send <- message:
				default:
					// close client send channel
					close(client.send)

					// delete client from map
					delete(hub.clients, client)
				}
			}
		}
	}
}
