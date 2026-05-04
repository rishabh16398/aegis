package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	pongTimeout    = 60 * time.Second
	pingInterval   = 54 * time.Second // must be less than pongTimeout
	maxMessageSize = 512
)

// Client represents one connected browser tab / frontend instance.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // outbound messages queued here
}

// readPump runs in its own goroutine.
// It reads messages from the WebSocket and handles disconnects.
// We don't expect the frontend to send us anything meaningful yet,
// but we must read to process control frames (ping/pong/close).
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	// Reset the deadline every time a pong arrives from the client.
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			// Normal close or network error — stop the loop.
			break
		}
	}
}

// writePump runs in its own goroutine.
// It drains the send channel and forwards each message to the WebSocket.
// It also sends periodic pings so we can detect dead connections.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				// Hub closed the channel — send a close frame.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				slog.Debug("ws write error", "err", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func eventToBytes(e Event) ([]byte, error) {
	return json.Marshal(e)
}
