package storage

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/marmyr/iagdbackup/internal/config"
	"time"
)

type CharacterDb struct {
}

type CharacterEntry struct {
	UserId    string    `json:"-" gorm:"column:userid"`
	Name      string    `json:"name" gorm:"column:name"`
	Filename  string    `json:"filename" gorm:"column:filename"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (CharacterEntry) TableName() string {
	return "character"
}

func (*CharacterDb) Get(user string, name string) (*CharacterEntry, error) {
	var entry CharacterEntry
	result := config.GetDatabaseInstance().Where("userid = ? AND name = ?", user, name).Take(&entry)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, result.Error
	}

	return &entry, result.Error
}

func (*CharacterDb) List(user string) ([]CharacterEntry, error) {
	DB := config.GetDatabaseInstance()

	var entries []CharacterEntry
	result := DB.Where("userid = ?", user).Find(&entries)

	return entries, result.Error
}

func (*CharacterDb) Insert(entry CharacterEntry) error {
	db := config.GetDatabaseInstance()

	result := db.Create(entry)
	if result.Error != nil {
		err := result.Error.(*pq.Error)
		if err.Code == UNIQUE_VIOLATION {
			return nil
		}
	}

	return result.Error
}

func (*CharacterDb) Purge(user string) error {
	db := config.GetDatabaseInstance()
	result := db.Where("userid = ?", user).Delete(CharacterEntry{})
	return result.Error
}
