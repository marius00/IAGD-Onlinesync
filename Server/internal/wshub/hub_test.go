package wshub

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// newTestServer spins up an httptest server that upgrades every connection and
// registers it with the given hub, echoing inbound messages back out via
// Broadcast (excluding the sender).
func newTestServer(t *testing.T, hub *Hub, email string) *httptest.Server {
	t.Helper()
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := up.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		client := hub.Register(email, conn)
		client.ReadLoop(func(msg []byte) {
			hub.Broadcast(email, client, msg)
		})
	}))
}

func dial(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + srv.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

func TestBroadcastReachesOtherClientButNotSender(t *testing.T) {
	hub := New()
	srv := newTestServer(t, hub, "user@example.com")
	defer srv.Close()

	sender := dial(t, srv)
	defer sender.Close()
	receiver := dial(t, srv)
	defer receiver.Close()

	// Give both registrations time to land.
	time.Sleep(50 * time.Millisecond)

	if err := sender.WriteMessage(websocket.TextMessage, []byte(`{"type":"item"}`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	receiver.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := receiver.ReadMessage()
	if err != nil {
		t.Fatalf("receiver did not get message: %v", err)
	}
	if string(msg) != `{"type":"item"}` {
		t.Fatalf("unexpected message: %s", msg)
	}

	// The sender must NOT receive its own broadcast.
	sender.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if _, _, err := sender.ReadMessage(); err == nil {
		t.Fatal("sender unexpectedly received its own broadcast")
	}
}

func TestBroadcastIsIsolatedPerEmail(t *testing.T) {
	hub := New()
	srvA := newTestServer(t, hub, "a@example.com")
	defer srvA.Close()
	srvB := newTestServer(t, hub, "b@example.com")
	defer srvB.Close()

	sender := dial(t, srvA)
	defer sender.Close()
	other := dial(t, srvB)
	defer other.Close()

	time.Sleep(50 * time.Millisecond)

	if err := sender.WriteMessage(websocket.TextMessage, []byte(`hello`)); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// A different e-mail must never receive another user's messages.
	other.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if _, _, err := other.ReadMessage(); err == nil {
		t.Fatal("client for a different e-mail received the message")
	}
}

func TestUnregisterRemovesClient(t *testing.T) {
	hub := New()
	srv := newTestServer(t, hub, "user@example.com")
	defer srv.Close()

	conn := dial(t, srv)
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	count := len(hub.clients["user@example.com"])
	hub.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 registered client, got %d", count)
	}

	conn.Close()
	// Allow the read loop to observe the close and unregister.
	time.Sleep(100 * time.Millisecond)

	hub.mu.RLock()
	_, exists := hub.clients["user@example.com"]
	hub.mu.RUnlock()
	if exists {
		t.Fatal("client was not unregistered after disconnect")
	}
}
