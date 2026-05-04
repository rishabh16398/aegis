package ws

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins during development.
	// In production this should check r.Header.Get("Origin").
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub is the central broadcaster.
// Any engine (scanner, network, process) calls hub.Publish(event)
// and the hub fans it out to every connected frontend client.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Event, 64), // buffered so publishers don't block
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run is the hub's event loop — call it in a goroutine from main.
// All map access is inside this single goroutine, so no mutex needed.
func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.clients[client] = true
			slog.Debug("ws client connected", "total", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				slog.Debug("ws client disconnected", "total", len(h.clients))
			}

		case event := <-h.broadcast:
			data, err := eventToBytes(event)
			if err != nil {
				slog.Error("ws marshal error", "err", err)
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					// Client's send buffer is full — it's too slow or dead.
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Publish sends an event to all connected clients.
// Safe to call from any goroutine.
func (h *Hub) Publish(e Event) {
	h.broadcast <- e
}

// ServeWS upgrades an HTTP request to a WebSocket connection
// and registers the new client with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "err", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	h.register <- client

	// Each client needs exactly two goroutines:
	// one for reading (detecting disconnects), one for writing (sending events).
	go client.writePump()
	go client.readPump()
}
