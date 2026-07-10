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

func (*CharacterDb) List(email string) ([]CharacterEntry, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	entries := make([]CharacterEntry, 0)
	err = db.Select(&entries, "SELECT name, filename FROM characters")
	return entries, err
}

func (*CharacterDb) Insert(email string, entry CharacterEntry) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO characters(name, filename) VALUES (?, ?)
		ON CONFLICT(name) DO UPDATE SET updated_at = unixepoch()`, entry.Name, entry.Filename)
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
