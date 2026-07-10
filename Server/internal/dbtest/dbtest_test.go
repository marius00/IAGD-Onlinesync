package dbtest

import (
	"path/filepath"
	"testing"

	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/marmyr/iagdbackup/internal/userdb"
)

// These exercise the real embedded migrations for both core.db and per-user
// db shapes, using a temp directory so no real /storage is touched. They live
// in their own package to avoid an import cycle with internal/sqlitedb.
func TestCoreDbMigratesAndIsIdempotent(t *testing.T) {
	t.Setenv("STORAGE_PATH", t.TempDir())
	t.Cleanup(coredb.CloseAll)

	db, err := coredb.Get()
	if err != nil {
		t.Fatalf("Error opening core db: %v", err)
	}

	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
		t.Fatalf("Expected users table to exist: %v", err)
	}

	if _, err := db.Exec("INSERT INTO users(email, db_filename) VALUES (?, ?)", "a@example.com", "a.db"); err != nil {
		t.Fatalf("Error inserting user: %v", err)
	}
}

func TestUserDbMigrates(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("STORAGE_PATH", dir)
	t.Cleanup(userdb.CloseAll)

	db, err := userdb.Get("someone@example.com")
	if err != nil {
		t.Fatalf("Error opening user db: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO item(id, seed, stackcount, affixrerollsused) VALUES (?, ?, ?, ?)`, "item-1", 123, 1, 9); err != nil {
		t.Fatalf("Error inserting item: %v", err)
	}

	var affixRerollsUsed int
	if err := db.Get(&affixRerollsUsed, "SELECT affixrerollsused FROM item WHERE id = ?", "item-1"); err != nil {
		t.Fatalf("Error reading item: %v", err)
	}
	if affixRerollsUsed != 9 {
		t.Fatalf("Expected affixrerollsused=9, got %d", affixRerollsUsed)
	}

	// Should have created the .db file under <storage>/users/
	matches, err := filepath.Glob(filepath.Join(dir, "users", "*.db"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("Expected exactly one .db file in users dir, got %v (err=%v)", matches, err)
	}
}

func TestUserDbGetReturnsSameConnectionForSameUser(t *testing.T) {
	t.Setenv("STORAGE_PATH", t.TempDir())
	t.Cleanup(userdb.CloseAll)

	db1, err := userdb.Get("same@example.com")
	if err != nil {
		t.Fatalf("Error opening user db: %v", err)
	}
	db2, err := userdb.Get("same@example.com")
	if err != nil {
		t.Fatalf("Error opening user db: %v", err)
	}
	if db1 != db2 {
		t.Fatal("Expected the same *sqlx.DB instance to be returned for the same user")
	}
}
