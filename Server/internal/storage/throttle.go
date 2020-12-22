package storage

import (
	"github.com/marmyr/myservice/internal/config"
	"time"
)

type ThrottleDb struct {
}

type ThrottleEntry struct {
	Id        int64     `json:"userid" gorm:"primaryKey"`
	UserId    string    `json:"userid" gorm:"column:userid"`
	Ip        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (ThrottleEntry) TableName() string {
	return "throttleentry"
}

// GetNumEntries returns the number of failed attempts by a user, in the past 2 hours
func (*ThrottleDb) GetNumEntries(user string, ip string) (int, error) {
	var entries []ThrottleEntry
	result := config.GetDatabaseInstance().Where("(userid = ? OR ip = ?) AND created_at > NOW() - INTERVAL '240 minutes'", user, ip).Find(&entries)

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

// Fetch all items queued to be deleted
func (*ThrottleDb) Purge(user string, ip string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ? OR ip = ?", user, ip).Delete(ThrottleEntry{})
	return result.Error
}

// Maintenance performs maintenance work such as deleting expired entries
func (*ThrottleDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	result := db.Where("created_at < NOW() - interval '1 day'").Delete(ThrottleEntry{})
	return result.Error
}