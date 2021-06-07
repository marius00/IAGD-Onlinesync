package mig

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"log"
	"time"
)

const MaxItemLimit = 10000

type PostgresOutputItem struct {
	UserId string `json:"-" gorm:"column:userid"`
	Id     string `json:"id"`
	Ts     int64  `json:"ts"`

	Mod        string `json:"mod"`
	IsHardcore bool   `json:"isHardcore" gorm:"column:ishardcore"`

	BaseRecord                 string `json:"baseRecord" gorm:"column:baserecord"`
	PrefixRecord               string `json:"prefixRecord" gorm:"column:prefixrecord"`
	SuffixRecord               string `json:"suffixRecord" gorm:"column:suffixrecord"`
	ModifierRecord             string `json:"modifierRecord" gorm:"column:modifierrecord"`
	TransmuteRecord            string `json:"transmuteRecord" gorm:"column:transmuterecord"`
	MateriaRecord              string `json:"materiaRecord" gorm:"column:materiarecord"`
	RelicCompletionBonusRecord string `json:"relicCompletionBonusRecord" gorm:"column:reliccompletionbonusrecord"`
	EnchantmentRecord          string `json:"enchantmentRecord" gorm:"column:enchantmentrecord"`

	// TODO: Buddy items does not need seed, but is it worth a new struct just to exclude it?
	Seed            int64 `json:"seed"`
	RelicSeed       int64 `json:"relicSeed" gorm:"column:relicseed"`
	EnchantmentSeed int64 `json:"enchantmentSeed" gorm:"column:enchantmentseed"`
	MateriaCombines int64 `json:"materiaCombines" gorm:"column:materiacombines"`
	StackCount      int64 `json:"stackCount" gorm:"column:stackcount"`

	// Used in IA for sorting/filtering
	CreatedAt int64 `json:"createdAt" gorm:"column:created_at"`

	// Metadata
	Name             string  `json:"name" gorm:"column:name"`
	NameLowercase    string  `json:"nameLowercase" gorm:"column:namelowercase"`
	Rarity           string  `json:"rarity" gorm:"column:rarity"`
	LevelRequirement float64 `json:"levelRequirement" gorm:"column:levelrequirement"`
	PrefixRarity     int64   `json:"prefixRarity" gorm:"column:prefixrarity"`
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func ListFromPostgres(lastTimestamp int64) ([]PostgresOutputItem, error) {
	DB := config.GetPostgresInstance()

	var items []PostgresOutputItem
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

func ToJsonItem(item PostgresOutputItem) storage.JsonItem {
	userDb := storage.UserDb{}
	user, err := userDb.GetByEmail(item.UserId)
	if err != nil {
		log.Fatalf("Error fetching user for item.. %v", err)
	}

	return storage.JsonItem{
		UserId:                     user.UserId,
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

func (PostgresOutputItem) TableName() string {
	return "item"
}

func InsertUser(entry storage.UserEntry) error {
	db := config.GetDatabaseInstance()
	result := db.Create(&entry)
	return result.Error
}
