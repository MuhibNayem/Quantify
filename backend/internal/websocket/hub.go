package websocket

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[uint]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[uint]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
		case client := <-h.unregister:
			if userClients, ok := h.clients[client.UserID]; ok {
				if _, ok := userClients[client]; ok {
					delete(userClients, client)
					close(client.Send)
					if len(userClients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
		case message := <-h.broadcast:
			for _, userClients := range h.clients {
				for client := range userClients {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(userClients, client)
					}
				}
			}
		}
	}
}

func (h *Hub) Broadcast(message interface{}) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Failed to marshal message to json: %v", err)
		return
	}
	h.broadcast <- jsonMessage
}

func (h *Hub) SendToUser(userID uint, message interface{}) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		logrus.Errorf("Failed to marshal message to json: %v", err)
		return
	}

	if userClients, ok := h.clients[userID]; ok {
		for client := range userClients {
			select {
			case client.Send <- jsonMessage:
			default:
				close(client.Send)
				delete(userClients, client)
			}
		}
	}
}
