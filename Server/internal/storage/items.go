package storage

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/util"
	"github.com/rs/zerolog/log"
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

/*
// Delete will delete an item for a user, both deleting the item row and making a "delete this item" row to signal other clients
func (self *ItemDb)  Delete(userId config.UserId, ids []string, timestamp int64) error {
	DB := config.GetDatabaseInstance()

	// Delete the actual items
	obj := InputItem{UserId: userId}

	query, args, err := sqlx.In("DELETE FROM items WHERE userId = ? AND id IN (?)", userId, ids)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query = DB.Rebind(query)
	_, err = DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Add deletion entries to sync deletes to other clients
	var deletionEntries []DeletedItem
	for _, id := range ids {
		deletionEntries = append(deletionEntries, DeletedItem{UserId: userId, Id: id, Ts: timestamp})
	}
	result = DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&deletionEntries)

	return ReturnOrIgnore(result.Error, UNIQUE_VIOLATION) // TODO: Unique viol not needed anymore?
}
*/
// Delete will delete an item for a user, both deleting the item row and making a "delete this item" row to signal other clients
func (self *ItemDb) Delete(ctx context.Context, userId config.UserId, ids []string, timestamp int64) error {
	db := config.GetDatabaseInstance()

	timeoutSeconds := time.Duration(2 * len(ids))
	timedCtx, cancel := context.WithTimeout(ctx, timeoutSeconds*time.Second)
	defer cancel()

	for _, id := range ids {
		ret, err := db.NamedExecContext(timedCtx, "DELETE FROM item WHERE userid = :userid AND id = :id", map[string]any{
			"userid": userId,
			"id":     id,
		})

		if err != nil {
			return err
		}

		rowsAffected, err := ret.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 1 {

			sql := "INSERT INTO deleteditem(userid, id, ts) VALUES (:userid, :id, :ts) ON DUPLICATE KEY UPDATE id=id"
			_, err = db.NamedExecContext(timedCtx, sql, map[string]interface{}{
				"userid": userId,
				"id":     id,
				"ts":     timestamp,
			})

			if err != nil {
				return err
			}
		} else {
			log.Warn().Msgf("Attempted to delete item %s, but item did not exist", id)
		}
	}

	return nil
}

// Maintenance deletes 'delete item' entries older than a year
func (self *ItemDb) Maintenance() error {
	db := config.GetDatabaseInstance()
	when := time.Now().AddDate(-1, 0, 0).Unix()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, "DELETE FROM deleteditem WHERE ts < ?", when)
	return err
}

func (self *ItemDb) Insert(userId config.UserId, items []InputItem) error {
	DB := config.GetDatabaseInstanceLegacy()
	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()

	for idx := range items {
		items[idx].UserId = userId

		// DB.NamedExecContext(ctx, "INSERT INTO items(id, userid, id_baserecord, id_prefixrecord, id_suffixrecord, id_modifierrecord, id_transmuterecord, seed, id_reliccompletionbonusrecord, id_enchantmentrecord, prefixrarity, unknown, enchantmentseed, materiacombines, stackcount, name, namelowercase, rarity, mod, levelrequirement, ishardcore, created_at, ts, relicseed, id_materiarecord) VALUES (:id, :userid, :base_record, :materia_record, :enchantment_record, :relic_completion_bonus_record, :transmute_record, :modifier_record, :suffix_record, :prefix_record, :mod, :prefix_rarity, :created_at, :enchantment_seed, :is_hardcore, :level_requirement, :materia_combines, :stack_count, :name, :name_lowercase, :rarity, :relic_seed, :seed, :ts)", items[idx])
	}

	result := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&items)
	return result.Error
}

// Fetch 0..1000 items for a given user, since the provided timestamp
func (self *ItemDb) List(user config.UserId, lastTimestamp int64) ([]OutputItem, error) {
	db := config.GetDatabaseInstanceLegacy()

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

// EnsureRecordsExists will insert any missing records for this item
func (self *ItemDb) ensureRecordsExists(items []JsonItem) error {

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
					if !RecordExists(record) {
						err := Write(record)
						if err != nil {
							log.Warn().Msgf("Failed to write record: %v", err)
						}
					}
					records = append(records, record)
				} else {
					fmt.Printf("Discarding record: %s\n", record)
				}
			}
		}
	}

	return nil
}

// Conerts a json item to an input item (settings record reference ids)
func (self *ItemDb) toInputItem(userId config.UserId, item JsonItem) InputItem {
	return InputItem{
		Id:                         item.Id,
		BaseRecord:                 ReadRecordId(item.BaseRecord),
		MateriaRecord:              ReadRecordId(item.MateriaRecord),
		EnchantmentRecord:          ReadRecordId(item.EnchantmentRecord),
		RelicCompletionBonusRecord: ReadRecordId(item.RelicCompletionBonusRecord),
		TransmuteRecord:            ReadRecordId(item.TransmuteRecord),
		ModifierRecord:             ReadRecordId(item.ModifierRecord),
		SuffixRecord:               ReadRecordId(item.SuffixRecord),
		PrefixRecord:               ReadRecordId(item.PrefixRecord),
		Mod:                        item.Mod,
		PrefixRarity:               item.PrefixRarity,
		CreatedAt:                  item.CreatedAt,
		EnchantmentSeed:            item.EnchantmentSeed,
		IsHardcore:                 item.IsHardcore,
		LevelRequirement:           item.LevelRequirement,
		MateriaCombines:            item.MateriaCombines,
		Name:                       item.Name,
		NameLowercase:              item.NameLowercase,
		Rarity:                     item.Rarity,
		RelicSeed:                  item.RelicSeed,
		Seed:                       item.Seed,
		StackCount:                 item.StackCount,
		Ts:                         item.Ts,
		UserId:                     userId,
	}
}

// Converts json items to input items, ensuring that the records exists in the database (mutates db)
func (self *ItemDb) ToInputItems(userId config.UserId, items []JsonItem) ([]InputItem, error) {
	if err := self.ensureRecordsExists(items); err != nil {
		return nil, err
	}

	var result []InputItem
	for _, item := range items {
		result = append(result, self.toInputItem(userId, item))
	}

	return result, nil
}

// testytest

// ListDeletedItems fetches all items queued to be deleted [a different client might have called delete, so it needs to sync down to all other clients]
func (self *ItemDb) ListDeletedItems(user config.UserId, lastTimestamp int64) ([]DeletedItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	DB := config.GetDatabaseInstance()

	args := map[string]any{
		"userid":    user,
		"timestamp": lastTimestamp,
	}
	var deletedItems []DeletedItem
	rows, err := DB.NamedQueryContext(ctx, "SELECT * FROM deleteditem WHERE userid = :userid AND ts > :timestamp", args)
	if err != nil {
		return deletedItems, err
	}

	for rows.Next() {
		var item DeletedItem
		err = rows.StructScan(&item)
		if err != nil {
			return deletedItems, err
		}

		deletedItems = append(deletedItems, item)
	}
	return deletedItems, nil
}

// Fetch all items queued to be deleted
func (self *ItemDb) Purge(user config.UserId) error {
	db := config.GetDatabaseInstance()

	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := db.ExecContext(ctx, "DELETE FROM item WHERE userid = ?", user)
		if err != nil {
			return err
		}
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := db.ExecContext(ctx, "DELETE FROM deleteditem WHERE userid = ?", user)
		if err != nil {
			return err
		}
	}

	return nil
}

func IsNotFoundError(err error) bool {
	return err != nil && err.Error() == gorm.ErrRecordNotFound.Error()
}
