package storage

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"time"
)

type ThrottleDb struct {
}

type ThrottleEntry struct {
	Id        int64     `json:"userid" gorm:"primaryKey"`
	UserId    string    `json:"-" gorm:"column:userid"`
	Ip        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (ThrottleEntry) TableName() string {
	return "throttleentry"
}

// GetNumEntries returns the number of failed attempts by a user, in the past 2 hours
func (*ThrottleDb) GetNumEntries(user string, ip string) (int, error) {
	var entries []ThrottleEntry
	result := config.GetDatabaseInstance().Where("(userid = ? OR ip = ?) AND created_at > NOW() - INTERVAL 240 minute", user, ip).Find(&entries)

	return len(entries), result.Error
}

func (*ThrottleDb) Insert(user string, ip string) error {
	DB := config.GetDatabaseInstance()

	result := DB.Create(&ThrottleEntry{
		UserId: user,
		Ip:     ip,
	})

	return result.Error
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

// Fetch all items queued to be deleted
func (*ThrottleDb) Purge(user string, ip string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ? OR ip = ?", user, ip).Delete(ThrottleEntry{})
	return result.Error
}

// Maintenance performs maintenance work such as deleting expired entries
func (*ThrottleDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	result := db.Where("created_at < NOW() - interval 1 day").Delete(ThrottleEntry{})
	return result.Error
}