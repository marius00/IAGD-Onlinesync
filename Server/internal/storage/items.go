package storage

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/marmyr/myservice/internal/config"
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

// Fetch all items in a partition for a given user
func (*ItemDb) List(user string, lastTimestamp int64) ([]Item, error) {
	DB := config.GetDatabaseInstance()

	var items []Item
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Find(&items)

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
