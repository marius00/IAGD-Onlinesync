package config

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// StoragePath is the mount point for all persisted SQLite databases (core.db and
// per-user .db files). On Coolify this is the host-mounted volume at /storage.
func StoragePath() string {
	if p := os.Getenv("STORAGE_PATH"); p != "" {
		return p
	}
	return "/storage"
}

// CoreDbPath returns the path to the shared core.db, which holds users, the
// records (string dedup) table, login attempts, throttle entries and migration
// bookkeeping.
func CoreDbPath() string {
	return filepath.Join(StoragePath(), "core.db")
}

// UserDbFilename returns the on-disk filename (not full path) for a given user's
// SQLite database: sha256(lowercased email) + ".db". Hashing avoids any issues
// with '@', unicode, path separators, or length limits present in raw e-mail
// addresses being used as filenames.
func UserDbFilename(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(email)))
	return hex.EncodeToString(sum[:]) + ".db"
}

// UserDbPath returns the full path to a given user's SQLite database under
// <StoragePath>/users/.
func UserDbPath(email string) string {
	return filepath.Join(StoragePath(), "users", UserDbFilename(email))
}
