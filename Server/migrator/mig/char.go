package mig

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"time"
)

func InsertCharactersToMysql(entry storage.CharacterEntry) error {
	db := config.GetDatabaseInstance()

	result :=

		db.Exec(`INSERT INTO characters(userid, name, filename, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE updated_at=now();`, entry.UserId, entry.Name, entry.Filename, entry.CreatedAt, entry.UpdatedAt)

	return result.Error
}

type PostgresCharacterEntry struct {
	Email     string    `json:"-" gorm:"column:userid"`
	Name      string    `json:"name" gorm:"column:name"`
	Filename  string    `json:"-" gorm:"column:filename"`
	CreatedAt time.Time `json:"createdAt" sql:"-" gorm:"-"`
	UpdatedAt time.Time `json:"updatedAt" sql:"-" gorm:"-"`
}

func (PostgresCharacterEntry) TableName() string {
	return "character"
}


func ListCharactersFromPostgres() ([]PostgresCharacterEntry, error) {
	DB := config.GetPostgresInstance()

	var entries []PostgresCharacterEntry
	result := DB.Find(&entries)

	return entries, result.Error
}