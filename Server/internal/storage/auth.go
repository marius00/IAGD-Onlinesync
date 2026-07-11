package storage

import (
	"database/sql"
	"errors"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"github.com/rs/zerolog/log"
	"time"
)

type AuthDb struct {
}

// AuthAttempt (login codes) lives in core.db: it's written for e-mails that
// may not have a user/db yet, so it can't be per-user.
type AuthAttempt struct {
	Key       string    `json:"key" db:"key"`
	Email     string    `json:"-" db:"email"`
	Code      string    `json:"-" db:"code"`
	Status    string    `json:"-" db:"status"` // Valid values: [COMPLETED, CREATED]
	CreatedAt time.Time `json:"created_at" db:"-"`
}

// GetUserId checks if an access token is valid for a given e-mail, returns 0 on invalid user/token combination.
//
// The user's own SQLite database is checked FIRST: tokens issued at login are
// written there (see StoreSuccessfulAuth), so a freshly-issued token must
// validate regardless of whether the user's historical data has finished
// draining out of MySQL. Only if the token isn't found in SQLite, and the user
// hasn't been migrated yet, do we fall back to their pre-existing tokens in the
// read-only MySQL source. This decoupling means login no longer depends on a
// successful (and possibly slow or failing) drain.
func (*AuthDb) GetUserId(email string, accessToken string) (config.UserId, error) {
	userDb := UserDb{}
	entry, err := userDb.GetByEmail(email)
	if err != nil {
		return 0, err
	}
	if entry == nil {
		return 0, nil
	}

	db, err := userdb.Get(email)
	if err != nil {
		return 0, err
	}

	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM authentry WHERE token = ?", accessToken); err != nil {
		return 0, err
	}
	if count > 0 {
		return entry.UserId, nil
	}

	// Not in SQLite: for users not yet drained, their older tokens still live in MySQL.
	if !IsMigrated(entry.UserId) && config.MySQLConfigured() {
		var mysqlCount int
		if err := config.GetDatabaseInstance().Get(&mysqlCount, "SELECT COUNT(*) FROM authentry WHERE userid = ? AND token = ?", entry.UserId, accessToken); err != nil {
			return 0, err
		}
		if mysqlCount > 0 {
			return entry.UserId, nil
		}
	}

	return 0, nil
}

// InitiateAuthentication initializes an authentication with key/code
func (*AuthDb) InitiateAuthentication(entry AuthAttempt) error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO authattempt(key, code, email, status) VALUES (?, ?, ?, ?)",
		entry.Key, entry.Code, entry.Email, entry.Status)
	return err
}

// Maintenance performs maintenance work such as deleting expired entries
func (*AuthDb) Maintenance() error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	since := time.Now().Add(-30 * time.Minute).Unix()
	_, err = db.Exec("DELETE FROM authattempt WHERE created_at < ?", since)
	return err
}

// GetAuthenticationAttempt fetches an auth attempt based on key and code
func (*AuthDb) GetAuthenticationAttempt(key string, code string) (*AuthAttempt, error) {
	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	since := time.Now().Add(-15 * time.Minute).Unix()

	var attempt AuthAttempt
	err = db.Get(&attempt, "SELECT key, code, email, status FROM authattempt WHERE key = ? AND code = ? AND created_at > ?", key, code, since)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &attempt, nil
}

func (*AuthDb) GetAuthenticationAttemptStatus(key string) (*AuthAttempt, error) {
	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	since := time.Now().Add(-15 * time.Minute).Unix()

	var attempt AuthAttempt
	err = db.Get(&attempt, "SELECT key, code, email, status FROM authattempt WHERE key = ? AND created_at > ?", key, since)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &attempt, nil
}

// GetLatestAuthToken returns the most recently issued access token for a user, or "" if none exists.
func (*AuthDb) GetLatestAuthToken(email string) string {
	db, err := userdb.Get(email)
	if err != nil {
		return ""
	}

	var token string
	if err := db.Get(&token, "SELECT token FROM authentry ORDER BY ts DESC LIMIT 1"); err != nil {
		return ""
	}

	return token
}

// StoreSuccessfulAuth stores an access token and marks the login attempt entry completed.
//
// The token is written and the attempt marked COMPLETED FIRST, so that login
// succeeds independently of the MySQL drain. The drain is then attempted on a
// best-effort basis: if it fails (e.g. flaky MySQL) or the request is cut off
// mid-drain, the user is still logged in with a valid token, and the drain is
// retried lazily on their next request and by the background drainer. Because
// the drain no longer clears the authentry table (see drainUser), a later drain
// will not wipe the token written here.
func (*AuthDb) StoreSuccessfulAuth(email string, userId config.UserId, key string, authToken string) error {
	userDb, err := userdb.Get(email)
	if err != nil {
		return err
	}

	if _, err := userDb.Exec("INSERT INTO authentry(token, ts) VALUES (?, ?)", authToken, time.Now().Unix()); err != nil {
		return err
	}

	coreDb, err := coredb.Get()
	if err != nil {
		return err
	}

	if _, err := coreDb.Exec("UPDATE authattempt SET status = 'COMPLETED' WHERE key = ?", key); err != nil {
		return err
	}

	// Best-effort: get the user's historical data into SQLite so it's ready on
	// first use. Never fail the login on this.
	if err := EnsureMigrated(email, userId); err != nil {
		log.Warn().Msgf("Login for %s succeeded, but initial MySQL drain failed (will retry later): %v", email, err)
	}

	return nil
}

// Purge will remove all access tokens and login attempts for the provided user
func (*AuthDb) Purge(user config.UserId, email string) error {
	coreDb, err := coredb.Get()
	if err != nil {
		return err
	}
	_, attemptErr := coreDb.Exec("DELETE FROM authattempt WHERE email = ?", email)

	userDb, err := userdb.Get(email)
	if err != nil {
		return err
	}
	if _, err := userDb.Exec("DELETE FROM authentry"); err != nil {
		return err
	}

	return attemptErr
}

// Logout removes a single access token for the provided user
func (*AuthDb) Logout(email string, accessToken string) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM authentry WHERE token = ?", accessToken)
	return err
}
