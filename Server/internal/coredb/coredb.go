// Package coredb provides access to the single shared core.db SQLite database:
// the user directory (email -> per-user db filename), the records (string
// dedup) table, login attempts, throttle entries, and migration bookkeeping.
package coredb

import (
	"embed"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/sqlitedb"
	"golang.org/x/sync/singleflight"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Cached by resolved path (rather than a bare sync.Once) so that tests which
// point STORAGE_PATH at different temp directories don't reuse a stale
// connection to a different core.db.
var (
	mu    sync.Mutex
	cache = map[string]*sqlx.DB{}
	group singleflight.Group
)

// Get returns the shared core.db handle, opening and migrating it on first
// use.
func Get() (*sqlx.DB, error) {
	path := config.CoreDbPath()

	mu.Lock()
	if db, ok := cache[path]; ok {
		mu.Unlock()
		return db, nil
	}
	mu.Unlock()

	db, err, _ := group.Do(path, func() (interface{}, error) {
		mu.Lock()
		if db, ok := cache[path]; ok {
			mu.Unlock()
			return db, nil
		}
		mu.Unlock()

		db, err := sqlitedb.Open(path, migrationsFS)
		if err != nil {
			return nil, err
		}

		mu.Lock()
		cache[path] = db
		mu.Unlock()
		return db, nil
	})

	if err != nil {
		return nil, err
	}
	return db.(*sqlx.DB), nil
}

// CloseAll closes and evicts every cached connection. Intended for graceful
// shutdown and for tests that need a temp directory to be removable
// afterwards (open SQLite file handles otherwise keep it locked on Windows).
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()

	for path, db := range cache {
		db.Close()
		delete(cache, path)
	}
}
