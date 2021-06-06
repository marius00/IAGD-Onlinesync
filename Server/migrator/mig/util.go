package mig

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"time"
)

const MaxItemLimit = 1000

// Fetch 0..1000 items for a given user, since the provided timestamp
func ListFromPostgres(lastTimestamp int64) ([]storage.OutputItem, error) {
	DB := config.GetPostgresInstance()

	var items []storage.OutputItem
	result := DB.Where("ts >= ?", lastTimestamp).Order("ts asc").Limit(MaxItemLimit).Find(&items)

	return items, result.Error
}

func ListDeletedItemsFromPostgres() ([]storage.DeletedItem, error) {
	DB := config.GetPostgresInstance()

	var items []storage.DeletedItem
	result := DB.Find(&items)

	return items, result.Error
}

func ResetItemDeletionInMysql() (error) {
	DB := config.GetDatabaseInstance()

	return DB.Raw("DELETE FROM deleteditem").Error
}

type PostgresUserEntry struct {
	UserId    string    `json:"-" gorm:"column:userid"`
	BuddyId   int32     `json:"buddyId" gorm:"column:buddy_id"`
	CreatedAt time.Time `json:"created_at" sql:"-" gorm:"-"`
}

func (PostgresUserEntry) TableName() string {
	return "users"
}

func ListUsersFromPostgres() ([]PostgresUserEntry, error) {
	DB := config.GetPostgresInstance()

	var users []PostgresUserEntry
	result := DB.Find(&users)

	return users, result.Error
}

func FindUser(email string, entries []storage.UserEntry) *storage.UserEntry {
	for _, user := range entries {
		if user.Email == email {
			return &user
		}
	}

	return nil
}

func FindUserP(email string, entries []PostgresUserEntry) *PostgresUserEntry {
	for _, user := range entries {
		if user.UserId == email {
			return &user
		}
	}

	return nil
}

func ListUsersFromMysql() ([]storage.UserEntry, error) {
	DB := config.GetDatabaseInstance()

	var users []storage.UserEntry
	result := DB.Find(&users)

	return users, result.Error
}

func ToJsonItem(item storage.OutputItem) storage.JsonItem {
	return storage.JsonItem{
		UserId:                     item.UserId,
		Id:                         item.Id,
		Ts:                         item.Ts,
		StackCount:                 item.StackCount,
		Seed:                       item.Seed,
		RelicSeed:                  item.RelicSeed,
		Rarity:                     item.Rarity,
		NameLowercase:              item.NameLowercase,
		Name:                       item.Name,
		MateriaCombines:            item.MateriaCombines,
		LevelRequirement:           item.LevelRequirement,
		IsHardcore:                 item.IsHardcore,
		EnchantmentSeed:            item.EnchantmentSeed,
		CreatedAt:                  item.CreatedAt,
		PrefixRarity:               item.PrefixRarity,
		Mod:                        item.Mod,
		PrefixRecord:               item.PrefixRecord,
		SuffixRecord:               item.SuffixRecord,
		ModifierRecord:             item.ModifierRecord,
		TransmuteRecord:            item.TransmuteRecord,
		RelicCompletionBonusRecord: item.RelicCompletionBonusRecord,
		EnchantmentRecord:          item.EnchantmentRecord,
		MateriaRecord:              item.MateriaRecord,
		BaseRecord:                 item.BaseRecord,
	}
}

func InsertUser(entry storage.UserEntry) error {
	db := config.GetDatabaseInstance()
	result := db.Create(entry)
	return result.Error
}
