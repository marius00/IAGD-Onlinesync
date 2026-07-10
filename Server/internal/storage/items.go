package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/marmyr/iagdbackup/internal/userdb"
	"github.com/marmyr/iagdbackup/internal/util"
	"github.com/rs/zerolog/log"
	"time"
)

const MaxItemLimit = 2500

type ItemDb struct {
}

// Delete will delete an item for a user, both deleting the item row and making a "delete this item" row to signal other clients
func (self *ItemDb) Delete(ctx context.Context, email string, ids []string, timestamp int64) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	timeoutSeconds := time.Duration(2 * len(ids))
	timedCtx, cancel := context.WithTimeout(ctx, timeoutSeconds*time.Second)
	defer cancel()

	for _, id := range ids {
		ret, err := db.ExecContext(timedCtx, "DELETE FROM item WHERE id = ?", id)
		if err != nil {
			return err
		}

		rowsAffected, err := ret.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 1 {
			_, err = db.ExecContext(timedCtx, "INSERT INTO deleteditem(id, ts) VALUES (?, ?) ON CONFLICT(id) DO NOTHING", id, timestamp)
			if err != nil {
				return err
			}
		} else {
			log.Warn().Msgf("Attempted to delete item %s, but item did not exist", id)
		}
	}

	return nil
}

// Maintenance deletes 'delete item' entries older than a year across all user databases.
func (self *ItemDb) Maintenance() error {
	userDb := UserDb{}
	emails, err := userDb.AllEmails()
	if err != nil {
		return err
	}

	when := time.Now().AddDate(-1, 0, 0).Unix()

	for _, email := range emails {
		db, err := userdb.Get(email)
		if err != nil {
			log.Warn().Msgf("Error opening db for %s during maintenance: %v", email, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = db.ExecContext(ctx, "DELETE FROM deleteditem WHERE ts < ?", when)
		cancel()
		if err != nil {
			log.Warn().Msgf("Error pruning deleted items for %s: %v", email, err)
		}
	}

	return nil
}

// insertColumns / insertPlaceholders keep the INSERT column list and its named
// bind placeholders in sync.
const insertColumns = `id, id_baserecord, id_prefixrecord, id_suffixrecord, id_modifierrecord,
	id_transmuterecord, id_reliccompletionbonusrecord, id_enchantmentrecord, id_materiarecord,
	id_ascendantaffixname, id_ascendantaffix2hname, seed, relicseed, enchantmentseed,
	materiacombines, stackcount, rerollsused, affixrerollsused, name, namelowercase, rarity,
	mod, levelrequirement, prefixrarity, ishardcore, created_at, ts`

const insertPlaceholders = `:id, :id_baserecord, :id_prefixrecord, :id_suffixrecord, :id_modifierrecord,
	:id_transmuterecord, :id_reliccompletionbonusrecord, :id_enchantmentrecord, :id_materiarecord,
	:id_ascendantaffixname, :id_ascendantaffix2hname, :seed, :relicseed, :enchantmentseed,
	:materiacombines, :stackcount, :rerollsused, :affixrerollsused, :name, :namelowercase, :rarity,
	:mod, :levelrequirement, :prefixrarity, :ishardcore, :created_at, :ts`

func (self *ItemDb) Insert(email string, items []InputItem) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO item (%s) VALUES (%s) ON CONFLICT(id) DO NOTHING", insertColumns, insertPlaceholders)

	for idx := range items {
		if _, err := db.NamedExec(query, items[idx]); err != nil {
			return err
		}
	}

	return nil
}

// outputRow scans the raw item row (numeric record ids) so records can be
// resolved in-process via the record cache, avoiding a multi-way JOIN.
type outputRow struct {
	Id string `db:"id"`
	Ts int64  `db:"ts"`

	Mod        string `db:"mod"`
	IsHardcore bool   `db:"ishardcore"`

	BaseRecord                 sql.NullInt64 `db:"id_baserecord"`
	PrefixRecord               sql.NullInt64 `db:"id_prefixrecord"`
	SuffixRecord               sql.NullInt64 `db:"id_suffixrecord"`
	ModifierRecord             sql.NullInt64 `db:"id_modifierrecord"`
	TransmuteRecord            sql.NullInt64 `db:"id_transmuterecord"`
	MateriaRecord              sql.NullInt64 `db:"id_materiarecord"`
	RelicCompletionBonusRecord sql.NullInt64 `db:"id_reliccompletionbonusrecord"`
	EnchantmentRecord          sql.NullInt64 `db:"id_enchantmentrecord"`
	AscendantAffixName         sql.NullInt64 `db:"id_ascendantaffixname"`
	AscendantAffix2hName       sql.NullInt64 `db:"id_ascendantaffix2hname"`

	Seed             int64 `db:"seed"`
	RelicSeed        int64 `db:"relicseed"`
	EnchantmentSeed  int64 `db:"enchantmentseed"`
	MateriaCombines  int64 `db:"materiacombines"`
	StackCount       int64 `db:"stackcount"`
	RerollsUsed      int64 `db:"rerollsused"`
	AffixRerollsUsed int64 `db:"affixrerollsused"`

	CreatedAt int64 `db:"created_at"`

	Name             string  `db:"name"`
	NameLowercase    string  `db:"namelowercase"`
	Rarity           string  `db:"rarity"`
	LevelRequirement float64 `db:"levelrequirement"`
	PrefixRarity     int64   `db:"prefixrarity"`
	Unknown          int64   `db:"unknown"`
}

func (r outputRow) toOutputItem() OutputItem {
	return OutputItem{
		Id:                         r.Id,
		Ts:                         r.Ts,
		Mod:                        r.Mod,
		IsHardcore:                 r.IsHardcore,
		BaseRecord:                 ReadRecord(r.BaseRecord),
		PrefixRecord:               ReadRecord(r.PrefixRecord),
		SuffixRecord:               ReadRecord(r.SuffixRecord),
		ModifierRecord:             ReadRecord(r.ModifierRecord),
		TransmuteRecord:            ReadRecord(r.TransmuteRecord),
		MateriaRecord:              ReadRecord(r.MateriaRecord),
		RelicCompletionBonusRecord: ReadRecord(r.RelicCompletionBonusRecord),
		EnchantmentRecord:          ReadRecord(r.EnchantmentRecord),
		AscendantAffixNameRecord:   ReadRecord(r.AscendantAffixName),
		AscendantAffix2hNameRecord: ReadRecord(r.AscendantAffix2hName),
		Seed:                       r.Seed,
		RelicSeed:                  r.RelicSeed,
		EnchantmentSeed:            r.EnchantmentSeed,
		MateriaCombines:            r.MateriaCombines,
		StackCount:                 r.StackCount,
		RerollsUsed:                r.RerollsUsed,
		AffixRerollsUsed:           r.AffixRerollsUsed,
		CreatedAt:                  r.CreatedAt,
		Name:                       r.Name,
		NameLowercase:              r.NameLowercase,
		Rarity:                     r.Rarity,
		LevelRequirement:           r.LevelRequirement,
		PrefixRarity:               r.PrefixRarity,
		Unknown:                    r.Unknown,
	}
}

const selectItemsQuery = `
SELECT id, ts, mod, ishardcore,
       id_baserecord, id_prefixrecord, id_suffixrecord, id_modifierrecord, id_transmuterecord,
       id_materiarecord, id_reliccompletionbonusrecord, id_enchantmentrecord,
       id_ascendantaffixname, id_ascendantaffix2hname,
       seed, relicseed, enchantmentseed, materiacombines, stackcount,
       IFNULL(rerollsused, 0) AS rerollsused, IFNULL(affixrerollsused, 0) AS affixrerollsused,
       name, namelowercase, rarity, levelrequirement, prefixrarity, IFNULL(unknown, 0) AS unknown,
       created_at
FROM item
WHERE ts > ?
ORDER BY ts ASC
LIMIT ?`

// List fetches 0..MaxItemLimit items for a given user, since the provided timestamp
func (self *ItemDb) List(ctx context.Context, email string, lastTimestamp int64) ([]OutputItem, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	timedCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.QueryxContext(timedCtx, selectItemsQuery, lastTimestamp, MaxItemLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items = make([]OutputItem, 0)
	for rows.Next() {
		var row outputRow
		if err := rows.StructScan(&row); err != nil {
			return nil, err
		}

		items = append(items, row.toOutputItem())
	}

	return items, nil
}

// ensureRecordsExists will insert any missing records for the given items into
// the shared records table (core.db) and update the in-memory cache.
func (self *ItemDb) ensureRecordsExists(items []JsonItem) error {
	for _, item := range items {
		candidates := []string{
			item.BaseRecord, item.PrefixRecord, item.SuffixRecord,
			item.ModifierRecord, item.TransmuteRecord, item.TransmuteRecord,
			item.EnchantmentRecord, item.MateriaRecord, item.AscendantAffix2hNameRecord, item.AscendantAffixNameRecord,
		}

		for _, record := range candidates {
			if record != "" {
				if util.IsASCII(record) {
					if !RecordExists(record) {
						if err := Write(record); err != nil {
							log.Warn().Msgf("Failed to write record: %v", err)
						}
					}
				} else {
					fmt.Printf("Discarding record: %s\n", record)
				}
			}
		}
	}

	return nil
}

// toInputItem converts a json item to an input item (resolving record reference ids)
func (self *ItemDb) toInputItem(item JsonItem) InputItem {
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
		AscendantAffixName:         ReadRecordId(item.AscendantAffixNameRecord),
		AscendantAffix2hName:       ReadRecordId(item.AscendantAffix2hNameRecord),
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
		RerollsUsed:                item.RerollsUsed,
		AffixRerollsUsed:           item.AffixRerollsUsed,
		Ts:                         item.Ts,
	}
}

// ToInputItems converts json items to input items, ensuring that the records exists in the shared records table (mutates core.db)
func (self *ItemDb) ToInputItems(items []JsonItem) ([]InputItem, error) {
	if err := self.ensureRecordsExists(items); err != nil {
		return nil, err
	}

	var result []InputItem
	for _, item := range items {
		result = append(result, self.toInputItem(item))
	}

	return result, nil
}

// ListDeletedItems fetches all items queued to be deleted [a different client might have called delete, so it needs to sync down to all other clients]
func (self *ItemDb) ListDeletedItems(email string, lastTimestamp int64) ([]DeletedItem, error) {
	db, err := userdb.Get(email)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var deletedItems = make([]DeletedItem, 0)
	rows, err := db.QueryxContext(ctx, "SELECT id, ts FROM deleteditem WHERE ts > ?", lastTimestamp)
	if err != nil {
		return deletedItems, err
	}
	defer rows.Close()

	for rows.Next() {
		var item DeletedItem
		if err := rows.StructScan(&item); err != nil {
			return deletedItems, err
		}

		deletedItems = append(deletedItems, item)
	}
	return deletedItems, nil
}

// Purge deletes all items and deletion markers for a user
func (self *ItemDb) Purge(email string) error {
	db, err := userdb.Get(email)
	if err != nil {
		return err
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_, err := db.ExecContext(ctx, "DELETE FROM item")
		cancel()
		if err != nil {
			return err
		}
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		_, err := db.ExecContext(ctx, "DELETE FROM deleteditem")
		cancel()
		if err != nil {
			return err
		}
	}

	return nil
}
