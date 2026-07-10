package session

import (
	"os"
	"testing"

	"github.com/marmyr/iagdbackup/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.IsolateStorage()
	os.Exit(m.Run())
}
