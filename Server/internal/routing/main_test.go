package routing

import (
	"os"
	"testing"

	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/internal/testutils"
)

// TestMain isolates SQLite state to a temp directory and pre-marks the dummy
// authorizer's user (userId 1, test@example.com) as migrated, so the auth
// middleware's EnsureMigrated call is a no-op rather than attempting a MySQL
// drain during the routing tests.
func TestMain(m *testing.M) {
	testutils.IsolateStorage()

	if err := storage.SetMigrated(1, 0, 0); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}
