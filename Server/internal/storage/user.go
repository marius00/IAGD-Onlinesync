package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/coredb"
	"math/rand"
	"strings"
	"time"
)

type UserDb struct {
}

type UserEntry struct {
	UserId     config.UserId `json:"-" db:"userid"`
	Email      string        `json:"-" db:"email"`
	BuddyId    int32         `json:"buddyId" db:"buddy_id"`
	DbFilename string        `json:"-" db:"db_filename"`
	CreatedAt  time.Time     `json:"created_at" db:"-"`
}

func (*UserDb) Get(user config.UserId) (*UserEntry, error) {
	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	var entry UserEntry
	err = db.Get(&entry, "SELECT userid, email, buddy_id, db_filename FROM users WHERE userid = ?", user)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (*UserDb) GetByEmail(email string) (*UserEntry, error) {
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("attempted to fetch user `%s` which is not a valid email", email)
	}

	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	var entry UserEntry
	err = db.Get(&entry, "SELECT userid, email, buddy_id, db_filename FROM users WHERE email = ?", email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

// GetFromBuddyId retrieves a UserEntry from the database identified by the given buddyId. Returns nil and no error if not found.
func (*UserDb) GetFromBuddyId(buddyId string) (*UserEntry, error) {
	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	var entry UserEntry
	err = db.Get(&entry, "SELECT userid, email, buddy_id, db_filename FROM users WHERE buddy_id = ?", buddyId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// Insert creates a user if one doesn't already exist for this email, and returns its userid either way.
func (*UserDb) Insert(entry UserEntry) (config.UserId, error) {
	db, err := coredb.Get()
	if err != nil {
		return config.UserId(0), err
	}

	entry.DbFilename = config.UserDbFilename(entry.Email)

	var created bool
	// Make up to N attempts to insert with a fresh random buddy_id (may conflict on buddy id).
	for i := 0; i < 100; i++ {
		entry.BuddyId = generateBuddyId()

		res, err := db.Exec("INSERT INTO users(email, buddy_id, db_filename) VALUES (?, ?, ?) ON CONFLICT(email) DO NOTHING",
			entry.Email, entry.BuddyId, entry.DbFilename)

		if err != nil {
			if isUniqueViolation(err) {
				// Either the email already exists (handled below regardless of loop outcome)
				// or the buddy_id collided (retry with a new one).
				continue
			}
			return config.UserId(0), err
		}

		if n, _ := res.RowsAffected(); n > 0 {
			created = true
		}
		break
	}

	existing, err := (&UserDb{}).GetByEmail(entry.Email)
	if err != nil {
		return config.UserId(0), err
	}
	if existing == nil {
		return config.UserId(0), errors.New("userid not returned")
	}

	// A brand-new user (not one bootstrapped from MySQL) has no legacy data to
	// drain, so mark them migrated immediately. This prevents a later drain from
	// clearing data written directly to their SQLite db (e.g. their first token).
	if created {
		if err := SetMigrated(existing.UserId, 0, 0); err != nil {
			return config.UserId(0), err
		}
	}

	return existing.UserId, nil
}

// AllEmails returns the e-mail of every registered user. Used by maintenance
// jobs that need to visit each user's database.
func (*UserDb) AllEmails() ([]string, error) {
	db, err := coredb.Get()
	if err != nil {
		return nil, err
	}

	var emails []string
	err = db.Select(&emails, "SELECT email FROM users")
	return emails, err
}

func (*UserDb) Purge(user config.UserId) error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM users WHERE userid = ?", user)
	return err
}

// init ensures that the random function is seeded at startup, so the pin codes are not generated in a predictable sequence.
func init() {
	rand.Seed(time.Now().Unix())
}

func generateBuddyId() int32 {
	return 100000 + rand.Int31n(99999)
}
