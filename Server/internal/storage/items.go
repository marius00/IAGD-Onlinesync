package storage

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/marmyr/iagdbackup/internal/config"
	"time"
)

const MaxItemLimit = 5000

type ItemDb struct {
}

const (
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
	UNIQUE_VIOLATION uint16 = 1062
)

// Delete will delete a an item for a user
func (*ItemDb) Delete(user string, id string, timestamp int64) error {
	DB := config.GetDatabaseInstance()

	obj := InputItem{Id: id, UserId: user}
	result := DB.Delete(&obj)
	if result.Error != nil && result.Error.Error() != gorm.ErrRecordNotFound.Error() {
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

func ReturnOrIgnore(err error, ignore uint16) error {
	if err != nil {
		err := err.(*mysql.MySQLError)
		if err.Number == ignore {
			return nil
		}
	}

	return err
}

func (*ItemDb) Insert(user string, item InputItem) error {
	DB := config.GetDatabaseInstance()

	item.UserId = user
	result := DB.Create(&item)
	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION)
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func (*ItemDb) List(user string, lastTimestamp int64) ([]OutputItem, error) {
	DB := config.GetDatabaseInstance()

	var items []OutputItem
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Order("ts asc").Limit(MaxItemLimit).Find(&items)

	return items, result.Error
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func insertRecordEntry(record string) error {
	DB := config.GetDatabaseInstance()
	result := DB.Exec("INSERT IGNORE INTO records(record) VALUES(?)", record)
	return result.Error
}

// EnsureRecordsExists will insert any missing records for this item
func (*ItemDb) ensureRecordsExists(items []JsonItem) {
	for _, item := range items {
		records := []string{
			item.BaseRecord, item.PrefixRecord, item.SuffixRecord,
			item.ModifierRecord, item.TransmuteRecord, item.TransmuteRecord,
			item.EnchantmentRecord, item.MateriaRecord,
		}

		for _, record := range records {
			if record != "" {
				insertRecordEntry(record)
			}
		}
	}
}

// Returns a string=>id map of the record references
func (*ItemDb) toMap(references []RecordReference) map[string]uint64 {
	var m map[string]uint64 = map[string]uint64{}
	for _, ref := range references {
		m[ref.Record] = ref.Id
	}

	return m
}

// Conerts a json item to an input item (settings record reference ids)
func (*ItemDb) toInputItem(item JsonItem, references map[string]uint64) InputItem {
	return InputItem{
		Id: item.Id,
		BaseRecord: references[item.BaseRecord],
		MateriaRecord: references[item.MateriaRecord],
		EnchantmentRecord: references[item.EnchantmentRecord],
		RelicCompletionBonusRecord: references[item.RelicCompletionBonusRecord],
		TransmuteRecord: references[item.TransmuteRecord],
		ModifierRecord: references[item.ModifierRecord],
		SuffixRecord: references[item.SuffixRecord],
		PrefixRecord: references[item.PrefixRecord],
		Mod: item.Mod,
		PrefixRarity: item.PrefixRarity,
		CreatedAt: item.CreatedAt,
		EnchantmentSeed: item.EnchantmentSeed,
		IsHardcore: item.IsHardcore,
		LevelRequirement: item.LevelRequirement,
		MateriaCombines: item.MateriaCombines,
		Name: item.Name,
		NameLowercase: item.NameLowercase,
		Rarity: item.Rarity,
		RelicSeed: item.RelicSeed,
		Seed: item.Seed,
		StackCount: item.StackCount,
		Ts: item.Ts,
		UserId: item.UserId,
	}
}

// Converts json items to input items, ensuring that the records exists in the database (mutates db)
func (db *ItemDb) ToInputItems(items []JsonItem) ([]InputItem, error) {
	db.ensureRecordsExists(items)
	ref, err := db.getRecordReferences(items)
	if err != nil {
		return nil, err
	}
	refMap := db.toMap(ref)

	var result []InputItem
	for _, item := range items {
		result = append(result, db.toInputItem(item, refMap));
	}

	return result, nil
}


// EnsureRecordsExists will insert any missing records for this item
func (*ItemDb) getRecordReferences(items []JsonItem) ([]RecordReference, error) {
	db := config.GetDatabaseInstance()
	var records []string
	for _, item := range items {
		for _, record := range []string{
			item.BaseRecord, item.PrefixRecord, item.SuffixRecord,
			item.ModifierRecord, item.TransmuteRecord, item.TransmuteRecord,
			item.EnchantmentRecord, item.MateriaRecord,
		} {
			if record != "" {
				records = append(records, record)
			}
		}
	}

	var references []RecordReference
	result := db.Where("record IN (?)", records).Find(&references)
	if result.Error != nil {
		return nil, result.Error
	}

	return references, nil
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

	result := db.Where("userid = ?", user).Delete(InputItem{})
	if result.Error != nil {
		return result.Error
	}

	result = db.Where("userid = ?", user).Delete(DeletedItem{})
	return result.Error
}


func IsNotFoundError(err error) bool {
	return err != nil && err.Error() != gorm.ErrRecordNotFound.Error()
}