package websocket

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// Hub maintains the set of active clients and broadcasts events to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Event
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Event, 256),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's event loop. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Debug().Int("clients", len(h.clients)).Msg("ws client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Debug().Int("clients", len(h.clients)).Msg("ws client disconnected")

		case event := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- event:
				default:
					// Slow client; drop and queue for cleanup
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()

		case <-h.done:
			return
		}
	}
}

func (h *Hub) Stop() {
	close(h.done)
}

// Broadcast queues an event for delivery to all connected clients.
func (h *Hub) Broadcast(event Event) {
	select {
	case h.broadcast <- event:
	default:
		log.Warn().Str("type", event.Type).Msg("ws broadcast channel full, dropping event")
	}
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
