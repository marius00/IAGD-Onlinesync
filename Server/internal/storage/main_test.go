package storage

import (
	"os"
	"testing"

	"github.com/marmyr/iagdbackup/internal/testutils"
)

// TestMain ensures tests never touch the production /storage mount: core.db
// and any per-user .db files created during this package's tests land under a
// throwaway temp directory. MySQL connection details still come from the
// environment (see the Makefile's `test` target, or docker-compose in
// D:\Dev\item for a local mirror of production).
func TestMain(m *testing.M) {
	testutils.IsolateStorage()
	os.Exit(m.Run())
}
