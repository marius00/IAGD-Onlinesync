package storage

import (
	"github.com/marmyr/myservice/internal/config"
	"time"
)

type UserDb struct {
}

type UserEntry struct {
	UserId    string    `json:"userid" gorm:"column:userid"`
	BuddyId   int32     `json:"buddyId" gorm:"column:buddy_id"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (UserEntry) TableName() string {
	return "users"
}

func (*UserDb) Get(user string) (*UserEntry, error) {
	var users []UserEntry
	result := config.GetDatabaseInstance().Where("userid = ?", user).Find(&users)
	if len(users) > 0 {
		return &users[0], result.Error
	}

	return nil, result.Error
}

func (*UserDb) Insert(entry UserEntry) error {
	db := config.GetDatabaseInstance()
	result := db.Create(entry)
	return result.Error
}

func (*UserDb) Purge(user string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ?", user).Delete(UserEntry{})
	return result.Error
}
