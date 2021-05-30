package main

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"gorm.io/gorm"
	"log"
)

// TODO: How did media types get unset in api gateway?
// TODO:
/*
[skip] Check if record exists
Insert into records - insert ignore
Insert into items (select wheeeere....)

: Ensure exists logic works
: Ensure test coverage on existing logic, /upload + /download maybe? To get the full flow.
: Then migrate.
*/

var db *gorm.DB

const MaxItemLimit = 1000

// Fetch 0..1000 items for a given user, since the provided timestamp
func listFromPostgres(lastTimestamp int64) ([]storage.OutputItem, error) {
	DB := config.GetPostgresInstance()

	var items []storage.OutputItem
	result := DB.Where("ts >= ?", lastTimestamp).Order("ts asc").Limit(MaxItemLimit).Find(&items)

	return items, result.Error
}

func toJsonItem(item storage.OutputItem) storage.JsonItem {
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

func main() {
	mysql := config.GetDatabaseInstance()

	// 1. Get max timestamp from mysql
	var highestTimestamp int64
	row := mysql.Table("item").Select("max(ts)", "", 0).Row()
	row.Scan(&highestTimestamp)

	// TODO: Insert and delete users

	var lastInsertedItems = map[string]struct{}{}
	for true {

		// Fetch batch of items
		postgresItems, err := listFromPostgres(highestTimestamp)
		if err != nil {
			log.Fatalf("Error fetching items from postgres, %v", err)
		}

		// Convert to json format
		var jsonItems []storage.JsonItem
		for _, item := range postgresItems {
			// Skip duplicates, will be some overlap between item batches
			_, exists := lastInsertedItems[item.Id]
			if exists {
				continue
			}

			jsonItems = append(jsonItems, toJsonItem(item))
		}

		// Convert to input format (and mutates mysql db, inserting records etc)
		itemDb := storage.ItemDb{}
		inputItems, err := itemDb.ToInputItems(jsonItems)
		if err != nil {
			log.Fatalf("Error converting items to InputItem, %v", err)
		}

		// Insert to mysql
		var currentlyInsertedItems = map[string]struct{}{}
		for _, item := range inputItems {

			if err = itemDb.Insert(item.UserId, item); err != nil {
				log.Fatalf("Error inserting items to mysql, %v", err)
			}

			if item.Ts > highestTimestamp {
				highestTimestamp = item.Ts
			}

			currentlyInsertedItems[item.Id] = struct{}{}
		}
		lastInsertedItems = currentlyInsertedItems

	}
	// TODO: Migrate DeletedItem & remove anything in DeletedItem
}
