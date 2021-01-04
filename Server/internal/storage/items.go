package storage

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/marmyr/myservice/internal/config"
	"strings"
	"time"
)

type ItemDb struct {
}

type Item struct {
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

	// TODO: Don't return this to IA, too much bloat
	SearchableText string `json:"searchableText" gorm:"column:searchabletext"`
	CachedStats    string `json:"cachedStats" gorm:"column:cachedstats"`
}

// We don't need to return all the stats, only a subset of the fields.
// Fields such as cached stats and searchable text are only used for the webview of backups
type OutputItem struct {
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
}
func (OutputItem) TableName() string {
	return "item"
}


type BuddyItem struct {
	UserId string `json:"-" gorm:"column:userid"`
	Id     string `json:"id"`
	CachedStats    string `json:"cachedStats" gorm:"column:cachedstats"`
}
func (BuddyItem) TableName() string {
	return "item"
}

type DeletedItem struct {
	UserId string `json:"-" gorm:"column:userid"`
	Id     string `json:"id"`
	Ts     int64  `json:"ts"`
}

func (DeletedItem) TableName() string {
	return "deleteditem"
}

const (
	// https://github.com/lib/pq/blob/master/error.go#L78
	UNIQUE_VIOLATION = "23505"
)

// Delete will delete a an item for a user
func (*ItemDb) Delete(user string, id string, timestamp int64) error {
	DB := config.GetDatabaseInstance()

	obj := Item{Id: id, UserId: user}
	result := DB.Delete(&obj)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	result = DB.Create(&DeletedItem{UserId: user, Id: id, Ts: timestamp})
	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION)
}

// Maintenance deletes 'delete item' entries older than a year
func (*ItemDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	when := time.Now().AddDate(-1, 0, 0)
	result := db.Where("ts < ?", when).Delete(DeletedItem{})
	return result.Error
}

func ReturnOrIgnore(err error, ignore pq.ErrorCode) error {
	if err != nil {
		err := err.(*pq.Error)
		if err.Code == ignore {
			return nil
		}
	}

	return err
}

func (*ItemDb) Insert(user string, item Item) error {
	DB := config.GetDatabaseInstance()

	item.UserId = user
	result := DB.Create(&item)
	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION)
}

// Fetch all items for a given user, since the provided timestamp
func (*ItemDb) List(user string, lastTimestamp int64) ([]OutputItem, error) {
	DB := config.GetDatabaseInstance()

	var items []OutputItem
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Find(&items)

	return items, result.Error
}

// Fetch all items for a given user, since the provided timestamp
func (*ItemDb) ListBuddyItems(user string, query []string, offset int64) ([]BuddyItem, error) {
	DB := config.GetDatabaseInstance()

	var items []BuddyItem

	var name = fmt.Sprintf("%%%s%%", strings.Join(query, " "))
	var db = DB.Where("userid = ?", user)
	for _, q := range query {
		db = db.Where("(searchabletext like ? OR namelowercase LIKE ?)", fmt.Sprintf("%%%s%%", q), name)
	}

	result := db.Order("name asc").Limit(35).Offset(offset).Find(&items)
	return items, result.Error
}

// Fetch all items queued to be deleted
func (*ItemDb) ListDeletedItems(user string, lastTimestamp int64) ([]DeletedItem, error) {
	DB := config.GetDatabaseInstance()

	var deletedItems []DeletedItem
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Find(&deletedItems)
	return deletedItems, result.Error
}

// Fetch all items queued to be deleted
func (*ItemDb) PurgeUser(user string) error {
	db := config.GetDatabaseInstance()

	result := db.Where("userid = ?", user).Delete(Item{})
	if result.Error != nil {
		return result.Error
	}

	result = db.Where("userid = ?", user).Delete(DeletedItem{})
	return result.Error
}
