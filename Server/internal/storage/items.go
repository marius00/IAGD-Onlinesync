package storage

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/util"
	"gorm.io/gorm/clause"
	"time"
)

const MaxItemLimit = 2500

type ItemDb struct {
}

const (
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
	UNIQUE_VIOLATION uint16 = 1062
)

// Delete will delete a an item for a user
func (*ItemDb) Delete(userId config.UserId, ids []string, timestamp int64) error {
	DB := config.GetDatabaseInstance()

	// Delete the actual items
	obj := InputItem{UserId: userId}
	result := DB.Where("userId = ? AND id IN (?)", userId, ids).Delete(&obj)
	if result.Error != nil && result.Error.Error() != gorm.ErrRecordNotFound.Error() {
		return result.Error
	}

	// Add deletion entries to sync deletes to other clients
	var deletionEntries []DeletedItem
	for _, id := range ids {
		deletionEntries = append(deletionEntries, DeletedItem{UserId: userId, Id: id, Ts: timestamp})
	}
	result = DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&deletionEntries)

	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION) // TODO: Unique viol not needed anymore?
}

// Maintenance deletes 'delete item' entries older than a year
func (*ItemDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	when := time.Now().AddDate(-1, 0, 0).Unix()
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

func (*ItemDb) Insert(userId config.UserId, items []InputItem) error {
	DB := config.GetDatabaseInstance()


	for idx := range items {
		items[idx].UserId = userId
	}

	result := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&items)
	return result.Error
}

func (*ItemDb) InsertOld(user config.UserId, item InputItem) error {
	DB := config.GetDatabaseInstance()

	item.UserId = user
	result := DB.Create(&item)
	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION)
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func (*ItemDb) List(user config.UserId, lastTimestamp int64) ([]OutputItem, error) {
	db := config.GetDatabaseInstance()

	sql := fmt.Sprintf(`
SELECT 
	id, 
	userid, 
	base.record AS baserecord,
	prefix.record as prefixrecord, 
	suffix.record as suffixrecord, 
	modifier.record as modifierrecord, 
	relic.record as reliccompletionbonusrecord,
	transmute.record as transmuterecord, 
	materia.record as materiarecord, 
	enchantment.record as enchantmentrecord, 
	seed,  
	relicseed, 
	prefixrarity, 
	unknown, 
	enchantmentseed, 
	materiacombines, 
	stackcount, 
	name, 
	namelowercase, 
	rarity, 
	levelrequirement, 
	%s, 
	ishardcore, 
	created_at, 
	ts
  FROM item i
  LEFT JOIN records as base ON i.id_baserecord = base.id_record
  LEFT JOIN records as prefix ON i.id_prefixrecord = prefix.id_record
  LEFT JOIN records AS suffix ON i.id_suffixrecord = suffix.id_record
  LEFT JOIN records AS modifier ON i.id_modifierrecord = modifier.id_record
  LEFT JOIN records AS transmute ON i.id_transmuterecord = transmute.id_record
  LEFT JOIN records AS materia ON i.id_materiarecord = materia.id_record
  LEFT JOIN records AS relic ON i.id_reliccompletionbonusrecord = relic.id_record
  LEFT JOIN records AS enchantment ON i.id_enchantmentrecord = enchantment.id_record
  WHERE userid = ? AND ts > ?
  ORDER BY ts ASC
  LIMIT ?
  `, "`mod`")
	var items = make([]OutputItem, 0)
	rows, err := db.Raw(sql, user, lastTimestamp, MaxItemLimit).Rows()
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var item OutputItem
	for rows.Next() {
		if err = db.ScanRows(rows, &item); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func insertRecordEntry(records []string) error {
	valueArgs := []interface{}{}
	sql := "INSERT IGNORE INTO records(record) VALUES"

	for _, val := range records {
		valueArgs = append(valueArgs, val)
		sql += "(?),"
	}


	// Remove the trailing ","
	sql = sql[0:len(sql)-1]

	DB := config.GetDatabaseInstance()
	result := DB.Exec(sql, valueArgs...)
	return result.Error
}

// EnsureRecordsExists will insert any missing records for this item
func (*ItemDb) ensureRecordsExists(items []JsonItem) error {

	var records []string

	for _, item := range items {
		candidates := []string{
			item.BaseRecord, item.PrefixRecord, item.SuffixRecord,
			item.ModifierRecord, item.TransmuteRecord, item.TransmuteRecord,
			item.EnchantmentRecord, item.MateriaRecord,
		}

		for _, record := range candidates {
			if record != "" {
				if util.IsASCII(record) {
					records = append(records, record)
				} else {
					fmt.Printf("Discarding record: %s\n", record)
				}
			}
		}
	}


	if len(records) > 0 {
		if err := insertRecordEntry(records); err != nil {
			return err
		}
	}

	return nil
}

// Returns a string=>id map of the record references
func (*ItemDb) toMap(references []RecordReference) map[string]sql.NullInt64 {

	var m = map[string]sql.NullInt64{
		"": {
			Valid:false,
		},
	}
	for _, ref := range references {
		m[ref.Record] = sql.NullInt64 { Int64: int64(ref.Id), Valid: true } // TODO: uint=>int cast, this will go to hell some day.
	}

	return m
}

// Conerts a json item to an input item (settings record reference ids)
func (*ItemDb) toInputItem(userId config.UserId, item JsonItem, references map[string]sql.NullInt64) InputItem {
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
		UserId: userId,
	}
}

// Converts json items to input items, ensuring that the records exists in the database (mutates db)
func (db *ItemDb) ToInputItems(userId config.UserId, items []JsonItem) ([]InputItem, error) {
	if err := db.ensureRecordsExists(items); err != nil {
		return nil, err
	}

	ref, err := db.getRecordReferences(items)
	if err != nil {
		return nil, err
	}
	refMap := db.toMap(ref)

	var result []InputItem
	for _, item := range items {
		result = append(result, db.toInputItem(userId, item, refMap));
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
			if record != "" && util.IsASCII(record) {
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
func (*ItemDb) ListDeletedItems(user config.UserId, lastTimestamp int64) ([]DeletedItem, error) {
	DB := config.GetDatabaseInstance()

	var deletedItems []DeletedItem
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Find(&deletedItems)
	return deletedItems, result.Error
}

// Fetch all items queued to be deleted
func (*ItemDb) Purge(user config.UserId) error {
	db := config.GetDatabaseInstance()

	result := db.Where("userid = ?", user).Delete(InputItem{})
	if result.Error != nil {
		return result.Error
	}

	result = db.Where("userid = ?", user).Delete(DeletedItem{})
	return result.Error
}


func IsNotFoundError(err error) bool {
	return err != nil && err.Error() == gorm.ErrRecordNotFound.Error()
}