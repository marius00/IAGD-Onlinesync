package storage

import (
	"database/sql"
	"errors"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"time"
)

type CharacterDb struct {
}

type CharacterEntry struct {
	Name      string    `json:"name" db:"name"`
	Filename  string    `json:"-" db:"filename"`
	CreatedAt time.Time `json:"createdAt" db:"-"`
	UpdatedAt time.Time `json:"updatedAt" db:"-"`
}

// CharacterSyncState is what the server already holds for a character: the
// digest of the stored archive and when it was last accepted. Character
// backups are a courtesy service, so both fields exist to avoid paying for an
// S3 upload we do not need.
type CharacterSyncState struct {
	ContentHash string `db:"content_hash"`
	UpdatedAt   int64  `db:"updated_at"`
}

// Age returns how long ago this character was last accepted.
func (s *CharacterSyncState) Age(now time.Time) time.Duration {
	return now.Sub(time.Unix(s.UpdatedAt, 0))
}

func (*CharacterDb) Get(email string, name string) (*CharacterEntry, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	var entry CharacterEntry
	err = db.Get(&entry, "SELECT name, filename FROM characters WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &entry, nil
}

// GetSyncState returns the stored digest and acceptance time for a character,
// or nil if this character has never been backed up.
func (*CharacterDb) GetSyncState(email string, name string) (*CharacterSyncState, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	var state CharacterSyncState
	err = db.Get(&state, "SELECT content_hash, updated_at FROM characters WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &state, nil
}

func (*CharacterDb) List(email string) ([]CharacterEntry, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	entries := make([]CharacterEntry, 0)
	err = db.Select(&entries, "SELECT name, filename FROM characters")
	return entries, err
}

// Upsert records a newly stored archive for a character.
func (*CharacterDb) Upsert(email string, entry CharacterEntry, contentHash string) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO characters(name, filename, content_hash) VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET filename = excluded.filename, content_hash = excluded.content_hash, updated_at = unixepoch()`,
		entry.Name, entry.Filename, contentHash)
	return err
}

// Touch bumps updated_at without replacing the stored archive. Used when a
// client re-uploads a character whose contents are byte-for-byte what we
// already hold, so the daily rate limit still counts the attempt.
func (*CharacterDb) Touch(email string, name string) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	_, err = db.Exec("UPDATE characters SET updated_at = unixepoch() WHERE name = ?", name)
	return err
}

func (*CharacterDb) Purge(email string) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM characters")
	return err
}
