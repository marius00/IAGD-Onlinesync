// Package userdb manages per-user SQLite databases: one file per user under
// <StoragePath>/users/<sha256(email)>.db, containing that user's items,
// deleted-item markers, character backups and access tokens.
//
// With ~23k users we cannot hold every file open at once, so connections are
// kept in a small LRU cache and opened on demand.
package userdb

import (
	"container/list"
	"embed"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/sqlitedb"
	"golang.org/x/sync/singleflight"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// maxOpenConnections bounds how many per-user .db files are held open at once.
// Least-recently-used connections beyond this are closed.
const maxOpenConnections = 200

type cacheEntry struct {
	path string
	db   *sqlx.DB
}

// Cached by resolved path (rather than bare filename) so that tests which
// point STORAGE_PATH at different temp directories don't reuse a stale
// connection to a different file.
var (
	mu      sync.Mutex
	entries = map[string]*list.Element{} // path -> lru element
	lru     = list.New()

	group singleflight.Group
)

// Get returns an open, migrated connection to the given user's database,
// identified by e-mail. Concurrent calls for the same user are serialized so
// only one connection is ever opened per user.
func Get(email string) (*sqlx.DB, error) {
	path := config.UserDbPath(email)

	if db, ok := lookup(path); ok {
		return db, nil
	}

	db, err, _ := group.Do(path, func() (interface{}, error) {
		if db, ok := lookup(path); ok {
			return db, nil
		}

		db, err := sqlitedb.Open(path, migrationsFS)
		if err != nil {
			return nil, err
		}

		store(path, db)
		return db, nil
	})

	if err != nil {
		return nil, err
	}
	return db.(*sqlx.DB), nil
}

func lookup(path string) (*sqlx.DB, bool) {
	mu.Lock()
	defer mu.Unlock()

	el, ok := entries[path]
	if !ok {
		return nil, false
	}

	lru.MoveToFront(el)
	return el.Value.(*cacheEntry).db, true
}

func store(path string, db *sqlx.DB) {
	mu.Lock()
	defer mu.Unlock()

	el := lru.PushFront(&cacheEntry{path: path, db: db})
	entries[path] = el

	for lru.Len() > maxOpenConnections {
		oldest := lru.Back()
		if oldest == nil {
			break
		}

		entry := oldest.Value.(*cacheEntry)
		lru.Remove(oldest)
		delete(entries, entry.path)
		entry.db.Close()
	}
}

// Remove closes any cached connection for the user and deletes their database
// file (plus WAL/SHM sidecars). Used on account deletion.
func Remove(email string) error {
	path := config.UserDbPath(email)

	mu.Lock()
	if el, ok := entries[path]; ok {
		el.Value.(*cacheEntry).db.Close()
		lru.Remove(el)
		delete(entries, path)
	}
	mu.Unlock()

	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}

// CloseAll closes and evicts every cached connection. Intended for graceful
// shutdown and for tests that need a temp directory to be removable
// afterwards (open SQLite file handles otherwise keep it locked on Windows).
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()

	for el := lru.Front(); el != nil; el = el.Next() {
		el.Value.(*cacheEntry).db.Close()
	}
	lru.Init()
	entries = map[string]*list.Element{}
}
