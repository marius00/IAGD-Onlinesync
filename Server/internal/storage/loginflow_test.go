package storage

import (
	"testing"

	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// TestTokenValidatesWhenNotMigrated is the regression test for the login bug:
// a token issued at login is written to the user's SQLite db, and must validate
// even if the user's historical MySQL data has NOT been drained yet (e.g. the
// drain failed or is still pending). Before the fix, GetUserId read tokens from
// MySQL for un-migrated users, so a fresh SQLite token returned 0 -> the user
// appeared logged out.
func TestTokenValidatesWhenNotMigrated(t *testing.T) {
	email := "loginfix-" + uuid.NewV4().String() + "@example.com"
	userId := config.UserId(870009001)

	core, err := coredb.Get()
	if err != nil {
		t.Fatalf("core.db: %v", err)
	}
	core.Exec("INSERT INTO users(userid, email, buddy_id, db_filename) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		userId, email, generateBuddyId(), config.UserDbFilename(email))
	t.Cleanup(func() {
		userdb.Remove(email)
		core.Exec("DELETE FROM users WHERE userid = ?", userId)
		core.Exec("DELETE FROM migration_state WHERE userid = ?", userId)
	})

	if IsMigrated(userId) {
		t.Skip("user unexpectedly already migrated")
	}

	// Simulate a login writing the token to SQLite without the user being migrated.
	udb, _ := userdb.Get(email)
	token := uuid.NewV4().String()
	if _, err := udb.Exec("INSERT INTO authentry(token, ts) VALUES (?, ?)", token, 1234567890); err != nil {
		t.Fatalf("insert token: %v", err)
	}

	gotUserId, err := (&AuthDb{}).GetUserId(email, token)
	assert.NoError(t, err)
	assert.Equalf(t, userId, gotUserId, "freshly-issued SQLite token must validate even when the user is not yet migrated")
}

// TestStoreSuccessfulAuthIssuesTokenAndMarksCompleted verifies the full happy
// path: the token is issued to SQLite, the attempt is marked COMPLETED, and the
// issued token validates.
func TestStoreSuccessfulAuthIssuesTokenAndMarksCompleted(t *testing.T) {
	email := "loginfix2-" + uuid.NewV4().String() + "@example.com"
	key := uuid.NewV4().String()
	code := "123456789"

	authDb := AuthDb{}
	userDb := UserDb{}

	assert.NoError(t, authDb.InitiateAuthentication(AuthAttempt{Email: email, Key: key, Code: code, Status: "CREATED"}))
	userId, err := userDb.Insert(UserEntry{Email: email})
	assert.NoError(t, err)
	t.Cleanup(func() {
		userdb.Remove(email)
		if core, e := coredb.Get(); e == nil {
			core.Exec("DELETE FROM users WHERE userid = ?", userId)
			core.Exec("DELETE FROM authattempt WHERE email = ?", email)
			core.Exec("DELETE FROM migration_state WHERE userid = ?", userId)
		}
	})

	token := uuid.NewV4().String()
	assert.NoError(t, authDb.StoreSuccessfulAuth(email, userId, key, token))

	// Attempt is COMPLETED and returns the issued token.
	status, err := authDb.GetAuthenticationAttemptStatus(key)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "COMPLETED", status.Status)
	assert.Equal(t, token, authDb.GetLatestAuthToken(email))

	// The issued token validates.
	gotUserId, err := authDb.GetUserId(email, token)
	assert.NoError(t, err)
	assert.Equal(t, userId, gotUserId)
}
