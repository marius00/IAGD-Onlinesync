package storage

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"time"
)

type CharacterDb struct {
}

type CharacterEntry struct {
	UserId    config.UserId `json:"-" gorm:"column:userid"`
	Name      string        `json:"name" gorm:"column:name"`
	Filename  string        `json:"-" gorm:"column:filename"`
	CreatedAt time.Time     `json:"createdAt" sql:"-" gorm:"-"`
	UpdatedAt time.Time     `json:"updatedAt" sql:"-" gorm:"-"`
}

func (CharacterEntry) Table() string {
	return "characters"
}
func (CharacterEntry) TableName() string {
	return "characters"
}

func (*CharacterDb) Get(user config.UserId, name string) (*CharacterEntry, error) {
	var entry CharacterEntry
	result := config.GetDatabaseInstanceLegacy().Where("userid = ? AND name = ?", user, name).Take(&entry)
	if result.Error != nil {
		if IsNotFoundError(result.Error) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &entry, result.Error
}

func (*CharacterDb) List(user config.UserId) ([]CharacterEntry, error) {
	DB := config.GetDatabaseInstanceLegacy()

	var entries []CharacterEntry
	result := DB.Where("userid = ?", user).Find(&entries)

	return entries, result.Error
}

func (*CharacterDb) Insert(entry CharacterEntry) error {
	db := config.GetDatabaseInstanceLegacy()

	result :=

		db.Exec(`INSERT INTO characters(userid, name, filename)
			VALUES(?, ?, ?)
			ON DUPLICATE KEY UPDATE updated_at=now();`, entry.UserId, entry.Name, entry.Filename)

	return result.Error
}

func (*CharacterDb) Purge(user config.UserId) error {
	db := config.GetDatabaseInstanceLegacy()
	result := db.Where("userid = ?", user).Delete(CharacterEntry{})
	return result.Error
}
