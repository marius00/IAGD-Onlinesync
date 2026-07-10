package storage

import (
	"github.com/marmyr/iagdbackup/internal/coredb"
	"time"
)

type ThrottleDb struct {
}

type ThrottleEntry struct {
	Id        int64     `json:"userid" db:"id"`
	UserId    string    `json:"-" db:"userid"`
	Ip        string    `json:"ip" db:"ip"`
	CreatedAt time.Time `json:"created_at" db:"-"`
}

// GetNumEntries returns the number of failed attempts by a user, in the past 4 hours
func (*ThrottleDb) GetNumEntries(user string, ip string) (int, error) {
	db, err := coredb.Get()
	if err != nil {
		return 0, err
	}

	since := time.Now().Add(-240 * time.Minute).Unix()

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM throttleentry WHERE (userid = ? OR ip = ?) AND created_at > ?", user, ip, since)
	return count, err
}

func (*ThrottleDb) Insert(user string, ip string) error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO throttleentry(userid, ip) VALUES (?, ?)", user, ip)
	return err
}

func (db *ThrottleDb) Throttle(user string, ip string, maxAttempts int) (bool, error) {
	numAttempts, err := db.GetNumEntries(user, ip)
	if err != nil {
		return true, err
	}

	if numAttempts > maxAttempts {
		return true, nil
	}

	if err := db.Insert(user, ip); err != nil {
		return true, err
	}

	return false, nil
}

// Purge removes all throttle entries for the given user or ip
func (*ThrottleDb) Purge(user string, ip string) error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM throttleentry WHERE userid = ? OR ip = ?", user, ip)
	return err
}

// Maintenance performs maintenance work such as deleting expired entries
func (*ThrottleDb) Maintenance() error {
	db, err := coredb.Get()
	if err != nil {
		return err
	}

	since := time.Now().Add(-24 * time.Hour).Unix()
	_, err = db.Exec("DELETE FROM throttleentry WHERE created_at < ?", since)
	return err
}
