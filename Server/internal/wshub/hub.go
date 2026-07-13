// Package wshub maintains the set of live WebSocket connections used to push
// item additions/deletions between the different PCs of a single user in near
// real-time.
//
// Connections are grouped per e-mail (a user logged in on multiple machines).
// The hub is a pure in-memory fan-out: it never persists anything itself, and
// it is safe for a connection to be dropped at any time -- the regular REST sync
// remains the source of truth and reconciles anything the hub missed.
//
// This is deliberately a single-process, in-memory design. The application runs
// as one Docker process, so that is sufficient. If it is ever scaled to multiple
// instances, fan-out would need to move to a shared pub/sub (e.g. Redis).
package wshub

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is how long a single write (or ping) may take before the
	// connection is considered dead.
	writeWait = 10 * time.Second

	// pongWait is how long we wait for a pong before assuming the peer is gone.
	pongWait = 60 * time.Second

	// pingPeriod must be less than pongWait; we ping at 90% of the pong window.
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize caps a single inbound message. Item/deletion batches are
	// limited to 100 entries client-side, so a few hundred KB is ample.
	maxMessageSize = 512 * 1024

	// sendBuffer bounds per-client outbound queue depth. A client that cannot
	// keep up is dropped rather than allowed to back the broadcaster up.
	sendBuffer = 64
)

// Client is a single live connection belonging to a user (identified by email).
type Client struct {
	hub   *Hub
	email string
	conn  *websocket.Conn
	send  chan []byte
	done  chan struct{}
	once  sync.Once
}

// Email returns the e-mail this connection is authenticated as.
func (c *Client) Email() string {
	return c.email
}

// close signals the write pump / read loop to stop. Safe to call repeatedly.
func (c *Client) close() {
	c.once.Do(func() { close(c.done) })
}

// trySend enqueues msg without blocking. Returns false if the buffer is full
// or the client is shutting down.
func (c *Client) trySend(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	case <-c.done:
		return false
	default:
		return false
	}
}

// writePump serializes all writes to the connection (gorilla requires a single
// writer) and keeps the connection alive with periodic pings.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

// ReadLoop reads and dispatches inbound messages until the connection closes.
// onMessage is invoked (synchronously) for each complete message. When it
// returns, the client is unregistered and the connection closed.
func (c *Client) ReadLoop(onMessage func(msg []byte)) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		onMessage(msg)
	}
}

// Hub tracks all live connections, grouped by e-mail.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
}

// New creates an empty hub.
func New() *Hub {
	return &Hub{clients: make(map[string]map[*Client]struct{})}
}

// Register adds a new connection for email and starts its write pump. The
// returned Client is used to drive the read loop and to exclude the sender from
// its own broadcasts.
func (h *Hub) Register(email string, conn *websocket.Conn) *Client {
	c := &Client{
		hub:   h,
		email: email,
		conn:  conn,
		send:  make(chan []byte, sendBuffer),
		done:  make(chan struct{}),
	}

	h.mu.Lock()
	if h.clients[email] == nil {
		h.clients[email] = make(map[*Client]struct{})
	}
	h.clients[email][c] = struct{}{}
	h.mu.Unlock()

	go c.writePump()
	return c
}

// Unregister removes a connection and stops its pumps. Safe to call more than
// once for the same client.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if set, ok := h.clients[c.email]; ok {
		if _, ok := set[c]; ok {
			delete(set, c)
			if len(set) == 0 {
				delete(h.clients, c.email)
			}
		}
	}
	h.mu.Unlock()

	c.close()
}

// Broadcast delivers msg to every connection for email except sender. Clients
// that cannot keep up are dropped; the REST sync will reconcile them later.
func (h *Hub) Broadcast(email string, sender *Client, msg []byte) {
	h.mu.RLock()
	var targets []*Client
	for c := range h.clients[email] {
		if c != sender {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range targets {
		if !c.trySend(msg) {
			c.close()
		}
	}
}
