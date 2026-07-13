// Package ws implements the /ws WebSocket endpoint used to push item additions
// and deletions between a user's machines in near real-time.
//
// Each message is both persisted (identically to the REST /upload and /remove
// endpoints) and relayed to the user's other connected machines. Persisting on
// receipt is important: the item's CloudId is assigned client-side at creation
// and travels with the message, so persisting it here means a later, idempotent
// REST upload of the same item is a no-op (ON CONFLICT(id) DO NOTHING) and peers
// can always deduplicate by that stable id.
package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/marmyr/iagdbackup/internal/logging"
	"github.com/marmyr/iagdbackup/internal/routing"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/util"
	"github.com/marmyr/iagdbackup/internal/wshub"
	"go.uber.org/zap"
)

const Path = "/ws"
const Method = routing.GET

// maxBatch mirrors the REST endpoints' per-request item cap.
const maxBatch = 100

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// The consumer is the IA desktop client, not a browser, so there is no
	// meaningful Origin to enforce.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// envelope is the wire format for a single WebSocket message. Exactly one of
// Items / Removed is populated, according to Type.
type envelope struct {
	Type    string             `json:"type"`
	Items   []storage.JsonItem `json:"items"`
	Removed []deleteEntry      `json:"removed"`
}

type deleteEntry struct {
	ID string `json:"id"`
}

// ProcessRequest returns the gin handler for the /ws endpoint. It must be
// mounted as a protected route so the auth middleware has already validated the
// token and populated the e-mail before the connection is upgraded.
func ProcessRequest(hub *wshub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logging.Logger(c)
		email := routing.GetEmail(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			// Upgrade already wrote an error response on failure.
			logger.Info("Failed to upgrade websocket connection", zap.Error(err), zap.String("user", email))
			return
		}

		logger.Info("Websocket connected", zap.String("user", email))
		client := hub.Register(email, conn)
		itemDb := &storage.ItemDb{}

		// ReadLoop blocks until the connection closes, so the gin handler (and
		// therefore the request/response writer) stays alive for its lifetime.
		client.ReadLoop(func(raw []byte) {
			var msg envelope
			if err := json.Unmarshal(raw, &msg); err != nil {
				logger.Info("Malformed websocket message", zap.Error(err), zap.String("user", email))
				return
			}

			switch msg.Type {
			case "item":
				if err := persistItems(itemDb, email, msg.Items); err != nil {
					logger.Warn("Failed to persist websocket items", zap.Error(err), zap.String("user", email))
					return
				}
			case "delete":
				if err := persistDeletions(itemDb, email, msg.Removed); err != nil {
					logger.Warn("Failed to persist websocket deletions", zap.Error(err), zap.String("user", email))
					return
				}
			default:
				logger.Info("Unknown websocket message type", zap.String("type", msg.Type), zap.String("user", email))
				return
			}

			// Fan the original message out to the user's other machines.
			hub.Broadcast(email, client, raw)
		})

		logger.Info("Websocket disconnected", zap.String("user", email))
	}
}

// persistItems validates and stores an item batch exactly as the REST /upload
// endpoint would.
func persistItems(db *storage.ItemDb, email string, items []storage.JsonItem) error {
	if err := validateItems(items); err != nil {
		return err
	}

	inputItems, err := db.ToInputItems(items)
	if err != nil {
		return err
	}

	ts := util.GetCurrentTimestamp()
	for idx := range inputItems {
		inputItems[idx].Ts = ts
	}

	return db.Insert(email, inputItems)
}

// persistDeletions validates and applies a deletion batch exactly as the REST
// /remove endpoint would (removes the item row and records a delete marker).
func persistDeletions(db *storage.ItemDb, email string, removed []deleteEntry) error {
	ids, err := validateDeletions(removed)
	if err != nil {
		return err
	}

	timeoutSeconds := time.Duration(2 * len(ids))
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSeconds*time.Second)
	defer cancel()

	return db.Delete(ctx, email, ids, util.GetCurrentTimestamp())
}

func validateItems(items []storage.JsonItem) error {
	if len(items) == 0 {
		return errInvalid("item batch is empty")
	}
	if len(items) > maxBatch {
		return errInvalid("item batch exceeds maximum size")
	}
	for _, m := range items {
		if len(m.Id) < 32 {
			return errInvalid(`item "id" must be of length 32 or longer`)
		}
		if m.Ts > 0 {
			return errInvalid(`item contains invalid property "_timestamp"`)
		}
		if len(m.BaseRecord) < 6 {
			return errInvalid(`item is missing the field "baseRecord"`)
		}
		if !m.HasValidRecords() {
			return errInvalid("item has one or more invalid records")
		}
		if m.HasOversizedString() {
			return errInvalid("item has a string field exceeding the maximum length")
		}
		if m.StackCount <= 0 {
			return errInvalid("item has a non-positive stack count")
		}
	}
	return nil
}

func validateDeletions(removed []deleteEntry) ([]string, error) {
	if len(removed) == 0 {
		return nil, errInvalid("deletion batch is empty")
	}
	if len(removed) > maxBatch {
		return nil, errInvalid("deletion batch exceeds maximum size")
	}
	ids := make([]string, 0, len(removed))
	for _, e := range removed {
		if len(e.ID) < 32 {
			return nil, errInvalid(`deletion "id" must be of length 32 or longer`)
		}
		ids = append(ids, e.ID)
	}
	return ids, nil
}

type validationError string

func (e validationError) Error() string { return string(e) }

func errInvalid(msg string) error { return validationError(msg) }
