package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/wshub"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// newServer starts a gin server exposing /ws behind a stub auth middleware that
// trusts the X-Api-User header (the real auth middleware is tested elsewhere).
// This lets each test client connect as an arbitrary e-mail.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	storage.Preload()

	hub := wshub.New()
	engine := gin.New()
	engine.GET("/ws", func(c *gin.Context) {
		email := strings.ToLower(c.GetHeader("X-Api-User"))
		c.Set("AuthEmailKey", email)
		c.Set("AuthUserKey", config.UserId(1))
		c.Next()
	}, ProcessRequest(hub))

	return httptest.NewServer(engine)
}

func dialAs(t *testing.T, srv *httptest.Server, email string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	header := http.Header{}
	header.Set("X-Api-User", email)
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return conn
}

func newEmail() string {
	return fmt.Sprintf("%s@example.com", uuid.NewV4().String())
}

func itemMessage(id, baseRecord string) []byte {
	env := map[string]interface{}{
		"type": "item",
		"items": []map[string]interface{}{
			{"id": id, "baseRecord": baseRecord, "stackCount": 1, "seed": 12345},
		},
	}
	raw, _ := json.Marshal(env)
	return raw
}

func deleteMessage(id string) []byte {
	env := map[string]interface{}{
		"type":    "delete",
		"removed": []map[string]interface{}{{"id": id}},
	}
	raw, _ := json.Marshal(env)
	return raw
}

func oversizedItemMessage(id string) []byte {
	env := map[string]interface{}{
		"type": "item",
		"items": []map[string]interface{}{
			{"id": id, "baseRecord": "my base record", "name": strings.Repeat("x", 256), "stackCount": 1},
		},
	}
	raw, _ := json.Marshal(env)
	return raw
}

func readWithin(conn *websocket.Conn, d time.Duration) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(d))
	_, msg, err := conn.ReadMessage()
	return msg, err
}

// A pushed item must be persisted to the sender's DB AND relayed to the peer.
func TestItemIsPersistedAndRelayed(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	email := newEmail()
	itemDb := storage.ItemDb{}
	defer itemDb.Purge(email)

	sender := dialAs(t, srv, email)
	defer sender.Close()
	receiver := dialAs(t, srv, email)
	defer receiver.Close()
	time.Sleep(50 * time.Millisecond) // let both registrations land

	id := uuid.NewV4().String()
	msg := itemMessage(id, "my base record")
	if err := sender.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Relayed to the peer verbatim.
	got, err := readWithin(receiver, 2*time.Second)
	assert.NoError(t, err, "receiver should get the relayed item")
	assert.JSONEq(t, string(msg), string(got))

	// Persisted server-side.
	items, err := itemDb.List(context.Background(), email, 0)
	assert.NoError(t, err)
	assert.Len(t, items, 1, "item should be persisted")
	if len(items) == 1 {
		assert.Equal(t, id, items[0].Id)
	}

	// The sender must not receive its own broadcast.
	_, err = readWithin(sender, 300*time.Millisecond)
	assert.Error(t, err, "sender should not receive its own broadcast")
}

// A pushed deletion must remove the item, record a delete marker, and be relayed.
func TestDeletionIsPersistedAndRelayed(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	email := newEmail()
	itemDb := storage.ItemDb{}
	defer itemDb.Purge(email)

	// Seed an item to delete.
	id := uuid.NewV4().String()
	input, err := itemDb.ToInputItems([]storage.JsonItem{{Id: id, BaseRecord: "my base record", StackCount: 1}})
	assert.NoError(t, err)
	assert.NoError(t, itemDb.Insert(email, input))

	sender := dialAs(t, srv, email)
	defer sender.Close()
	receiver := dialAs(t, srv, email)
	defer receiver.Close()
	time.Sleep(50 * time.Millisecond)

	msg := deleteMessage(id)
	if err := sender.WriteMessage(websocket.TextMessage, msg); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	got, err := readWithin(receiver, 2*time.Second)
	assert.NoError(t, err, "receiver should get the relayed deletion")
	assert.JSONEq(t, string(msg), string(got))

	items, err := itemDb.List(context.Background(), email, 0)
	assert.NoError(t, err)
	assert.Len(t, items, 0, "item should be deleted")

	deleted, err := itemDb.ListDeletedItems(email, 0)
	assert.NoError(t, err)
	assert.Len(t, deleted, 1, "a delete marker should be recorded")
}

// Fan-out must be scoped to a single e-mail.
func TestBroadcastIsIsolatedPerEmail(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	emailA := newEmail()
	emailB := newEmail()
	itemDb := storage.ItemDb{}
	defer itemDb.Purge(emailA)

	sender := dialAs(t, srv, emailA)
	defer sender.Close()
	other := dialAs(t, srv, emailB)
	defer other.Close()
	time.Sleep(50 * time.Millisecond)

	if err := sender.WriteMessage(websocket.TextMessage, itemMessage(uuid.NewV4().String(), "my base record")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := readWithin(other, 400*time.Millisecond)
	assert.Error(t, err, "a different e-mail must not receive the message")
}

// An item with an oversized string field must be neither persisted nor relayed.
func TestOversizedStringIsRejected(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	email := newEmail()
	itemDb := storage.ItemDb{}
	defer itemDb.Purge(email)

	sender := dialAs(t, srv, email)
	defer sender.Close()
	receiver := dialAs(t, srv, email)
	defer receiver.Close()
	time.Sleep(50 * time.Millisecond)

	if err := sender.WriteMessage(websocket.TextMessage, oversizedItemMessage(uuid.NewV4().String())); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := readWithin(receiver, 400*time.Millisecond)
	assert.Error(t, err, "oversized item must not be relayed")

	items, err := itemDb.List(context.Background(), email, 0)
	assert.NoError(t, err)
	assert.Len(t, items, 0, "oversized item must not be persisted")
}

// An invalid message must be neither persisted nor relayed.
func TestInvalidItemIsRejected(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	email := newEmail()
	itemDb := storage.ItemDb{}
	defer itemDb.Purge(email)

	sender := dialAs(t, srv, email)
	defer sender.Close()
	receiver := dialAs(t, srv, email)
	defer receiver.Close()
	time.Sleep(50 * time.Millisecond)

	// baseRecord too short -> fails validation.
	if err := sender.WriteMessage(websocket.TextMessage, itemMessage(uuid.NewV4().String(), "no")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := readWithin(receiver, 400*time.Millisecond)
	assert.Error(t, err, "invalid item must not be relayed")

	items, err := itemDb.List(context.Background(), email, 0)
	assert.NoError(t, err)
	assert.Len(t, items, 0, "invalid item must not be persisted")
}
