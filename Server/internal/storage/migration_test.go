package storage

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// TestDrainRealUserFromMySQL drains an actual user (one that has items) from the
// local MySQL mirror into a per-user SQLite db and verifies the item counts
// match. Requires the local MySQL mirror (see D:\Dev\item).
func TestDrainRealUserFromMySQL(t *testing.T) {
	if !config.MySQLConfigured() {
		t.Skip("MySQL not configured")
	}

	mysql := config.GetDatabaseInstance()

	var picked struct {
		UserId config.UserId `db:"userid"`
		Email  string        `db:"email"`
		Count  int           `db:"c"`
	}
	// Pick a userid straight from the item table (cheap), then look up its email
	// and exact item count. Avoids a GROUP BY over the whole 15M-row table.
	if err := mysql.Get(&picked.UserId, "SELECT userid FROM item LIMIT 1"); err != nil {
		t.Skipf("No MySQL items available: %v", err)
	}
	if err := mysql.Get(&picked.Email, "SELECT email FROM users WHERE userid = ?", picked.UserId); err != nil {
		t.Skipf("Picked item's user not present in users table: %v", err)
	}
	if err := mysql.Get(&picked.Count, "SELECT COUNT(*) FROM item WHERE userid = ?", picked.UserId); err != nil {
		t.Fatalf("Error counting items: %v", err)
	}

	// Register the user in core.db (preserving userid), as bootstrap would.
	core, err := coredb.Get()
	if err != nil {
		t.Fatalf("Error opening core.db: %v", err)
	}
	_, err = core.Exec("INSERT INTO users(userid, email, buddy_id, db_filename) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		picked.UserId, picked.Email, generateBuddyId(), config.UserDbFilename(picked.Email))
	assert.NoErrorf(t, err, "Error registering user in core.db")

	t.Cleanup(func() {
		userdb.Remove(picked.Email)
		core.Exec("DELETE FROM users WHERE userid = ?", picked.UserId)
		core.Exec("DELETE FROM migration_state WHERE userid = ?", picked.UserId)
	})

	assert.False(t, IsMigrated(picked.UserId), "User should not be migrated yet")

	err = EnsureMigrated(picked.Email, picked.UserId)
	assert.NoErrorf(t, err, "Error draining user")
	assert.True(t, IsMigrated(picked.UserId), "User should be migrated after drain")

	// Verify the SQLite db now holds exactly the MySQL item count.
	udb, err := userdb.Get(picked.Email)
	assert.NoErrorf(t, err, "Error opening user db")

	var sqliteCount int
	err = udb.Get(&sqliteCount, "SELECT COUNT(*) FROM item")
	assert.NoErrorf(t, err, "Error counting items")
	assert.Equalf(t, picked.Count, sqliteCount, "Expected SQLite item count to match MySQL")

	// EnsureMigrated is idempotent.
	err = EnsureMigrated(picked.Email, picked.UserId)
	assert.NoErrorf(t, err, "Second EnsureMigrated should be a no-op")

	// A subsequent List should succeed against the drained data.
	Preload()
	itemDb := ItemDb{}
	_, err = itemDb.List(context.Background(), picked.Email, 0)
	assert.NoErrorf(t, err, "Error listing drained items")
}

// TestDrainSkipsCorruptRows verifies that individual corrupt item rows are
// skipped rather than failing the whole user's migration. A row with an
// id_baserecord larger than int64 (valid in MySQL's bigint UNSIGNED, but
// unscannable into a Go int64) stands in for the known prod corruption.
func TestDrainSkipsCorruptRows(t *testing.T) {
	if !config.MySQLConfigured() {
		t.Skip("MySQL not configured")
	}

	mysql := config.GetDatabaseInstance()

	userId := config.UserId(990000001)
	email := fmt.Sprintf("corrupt-%s@example.com", uuid.NewV4().String())

	// Seed a synthetic user with 2 good rows and 1 corrupt row in MySQL. FK checks
	// are disabled so we don't need matching records/users rows for this fixture.
	tx := mysql.MustBegin()
	tx.Exec("SET FOREIGN_KEY_CHECKS=0")
	tx.Exec("INSERT INTO users(userid, email) VALUES (?, ?)", userId, email)
	good1 := uuid.NewV4().String()
	good2 := uuid.NewV4().String()
	corrupt := uuid.NewV4().String()
	tx.Exec("INSERT INTO item(id, userid, id_baserecord, seed, stackcount, created_at, ts) VALUES (?, ?, 1, 100, 1, 0, 10)", good1, userId)
	tx.Exec("INSERT INTO item(id, userid, id_baserecord, seed, stackcount, created_at, ts) VALUES (?, ?, 1, 100, 1, 0, 10)", good2, userId)
	// 18446744073709551615 = max uint64, unscannable into int64 -> row is skipped.
	tx.Exec("INSERT INTO item(id, userid, id_baserecord, seed, stackcount, created_at, ts) VALUES (?, ?, 18446744073709551615, 100, 1, 0, 10)", corrupt, userId)
	if err := tx.Commit(); err != nil {
		t.Fatalf("Error seeding corrupt fixture: %v", err)
	}

	core, err := coredb.Get()
	if err != nil {
		t.Fatalf("Error opening core.db: %v", err)
	}
	core.Exec("INSERT INTO users(userid, email, buddy_id, db_filename) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		userId, email, generateBuddyId(), config.UserDbFilename(email))

	t.Cleanup(func() {
		mysql.Exec("DELETE FROM item WHERE userid = ?", userId)
		mysql.Exec("DELETE FROM users WHERE userid = ?", userId)
		core.Exec("DELETE FROM users WHERE userid = ?", userId)
		core.Exec("DELETE FROM migration_state WHERE userid = ?", userId)
		userdb.Remove(email)
	})

	// Drain should succeed (not fail on the corrupt row) and migrate the 2 good rows.
	err = EnsureMigrated(email, userId)
	assert.NoErrorf(t, err, "Drain should tolerate the corrupt row")
	assert.True(t, IsMigrated(userId), "User should be marked migrated")

	udb, _ := userdb.Get(email)
	var count int
	udb.Get(&count, "SELECT COUNT(*) FROM item")
	assert.Equalf(t, 2, count, "Expected the 2 good rows migrated, corrupt row skipped")
}

// TestDrainExhaustion verifies the FAILED/attempt bookkeeping so the background
// drainer gives up on a chronically failing user instead of getting stuck.
func TestDrainExhaustion(t *testing.T) {
	userId := config.UserId(990000002)

	t.Cleanup(func() {
		if core, err := coredb.Get(); err == nil {
			core.Exec("DELETE FROM migration_state WHERE userid = ?", userId)
		}
	})

	assert.False(t, isDrainExhausted(userId), "Fresh user should not be exhausted")

	for i := 0; i < maxDrainAttempts; i++ {
		markFailed(userId, errors.New("boom"))
	}

	assert.True(t, isDrainExhausted(userId), "User should be exhausted after maxDrainAttempts")

	core, _ := coredb.Get()
	var st struct {
		Status   string `db:"status"`
		Attempts int    `db:"attempts"`
	}
	err := core.Get(&st, "SELECT status, attempts FROM migration_state WHERE userid = ?", userId)
	assert.NoErrorf(t, err, "Expected a migration_state row")
	assert.Equal(t, "FAILED", st.Status)
	assert.GreaterOrEqual(t, st.Attempts, maxDrainAttempts)

	// A later success clears the failure state.
	assert.NoError(t, SetMigrated(userId, 3, 3))
	assert.False(t, isDrainExhausted(userId), "Successful drain should clear exhaustion")
	assert.True(t, IsMigrated(userId), "User should now be migrated")
}

// TestSetMigratedMarksInMemoryEvenOnPersistFailure is a regression test for a
// bug where a failure to persist migration_state to core.db (while the data
// itself was already safely committed to the user's SQLite db) left the
// in-memory migrated flag unset. That caused the next EnsureMigrated call to
// re-run the clear-then-copy drain, silently wiping anything written to the
// user's db since (e.g. a freshly-issued access token written immediately
// after StoreSuccessfulAuth's own EnsureMigrated call).
func TestSetMigratedMarksInMemoryEvenOnPersistFailure(t *testing.T) {
	core, err := coredb.Get()
	if err != nil {
		t.Fatalf("Error opening core.db: %v", err)
	}

	// Force the core.db write inside SetMigrated to fail, without touching the
	// in-memory state we're trying to verify.
	if _, err := core.Exec("ALTER TABLE migration_state RENAME TO migration_state_test_bak"); err != nil {
		t.Fatalf("Error renaming migration_state for test: %v", err)
	}
	t.Cleanup(func() {
		core.Exec("DROP TABLE IF EXISTS migration_state")
		core.Exec("ALTER TABLE migration_state_test_bak RENAME TO migration_state")
	})

	userId := config.UserId(990000003)

	err = SetMigrated(userId, 5, 5)
	assert.Error(t, err, "Expected the core.db write to fail with the table renamed away")
	assert.True(t, IsMigrated(userId), "User must be marked migrated in-memory even though the durable record failed to write")
}
