package main

import (
	"github.com/marmyr/iagdbackup/internal/config"
	"github.com/marmyr/iagdbackup/internal/storage"
	"github.com/marmyr/iagdbackup/migrator/mig"
	"gorm.io/gorm"
	"log"
	"time"
)

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



type PostgresAuthEntry struct {
	UserId string    `json:"-" gorm:"column:userid"`
	Token  string    `json:"-"`
	Ts     time.Time `json:"ts"`
}
func (PostgresAuthEntry) TableName() string {
	return "authentry"
}

func listPostgresAuthTokens() ([]PostgresAuthEntry, error) {
	DB := config.GetPostgresInstance()

	var tokens []PostgresAuthEntry
	result := DB.Find(&tokens)

	return tokens, result.Error
}

func storeAuthTokens(tokens []PostgresAuthEntry) {
	DB := config.GetDatabaseInstance()
	// TODO: Can't do this, wiping auth tokens in prod is bad.. forced logouts!
	/*if err :=DB.Raw("DELETE FROM authentry").Where("1 = ?", 1).Error; err != nil {
		log.Fatalf("Error deleting auth entries, %v", err)
	}*/

	for _, entry := range tokens {
		if err := storage.ReturnOrIgnore(DB.Create(&storage.AuthEntry{
			Email: entry.UserId,
			Ts: entry.Ts,
			Token: entry.Token,
			UserId: mig.GetUserByEmail(entry.UserId).UserId,
		}).Error, storage.UNIQUE_VIOLATION); err != nil {
			log.Fatalf("Error inserting auth token, %v", err)
		}
	}
}

// Migrate users over to mysql
func migrateUsers() {
	itemDb := storage.ItemDb{}
	authDb := storage.AuthDb{}
	characterDb := storage.CharacterDb{}


	postgresUsers, err := mig.ListUsersFromPostgres()
	if err != nil {
		log.Fatalf("Error fetching users from postgres, %v", err)
	}

	mysqlUsers, err := mig.ListUsersFromMysql()
	if err != nil {
		log.Fatalf("Error fetching users from mysql, %v", err)
	}

	log.Println("Inserting missing users")
	userDb := storage.UserDb{}
	for _, user := range postgresUsers {
		// If it doesn't exist in mysql, insert it.
		if mig.FindUser(user.UserId, mysqlUsers) == nil {
			log.Printf("Inserting user %s\n", user.UserId)
			if err = mig.InsertUser(storage.UserEntry{
				Email: user.UserId,
				CreatedAt: user.CreatedAt,
				BuddyId: user.BuddyId,
			}); err != nil {
				log.Fatalf("Unabled to insert user %v", err)
			}
			// TODO: Merge auth tokens
		}
	}

	authTokens, err := listPostgresAuthTokens()
	if err != nil {
		log.Fatalf("Error fetching auth tokens, %v", err)
	}
	storeAuthTokens(authTokens)


	log.Println("Deleting purged users")
	for _, user := range mysqlUsers {
		// If it doesn't exist in postgres, delete it.
		if mig.FindUserP(user.Email, postgresUsers) == nil {
			log.Printf("Purging user %v\n", user.UserId)
			if err = itemDb.Purge(user.UserId); err != nil {
				log.Fatalf("Unabled to purge items for user %v", err)
			}
			if err = authDb.Purge(user.UserId, user.Email); err != nil {
				log.Fatalf("Unabled to purge auth token for user %v", err)
			}

			if err = characterDb.Purge(user.UserId); err != nil {
				log.Fatalf("Unabled to purge characters for user %v", err)
			}

			if err = userDb.Purge(user.UserId); err != nil {
				log.Fatalf("Unabled to purge user %v", err)
			}
		}
	}
}


func getItemBatch(highestTimestamp int64, lastInsertedItems map[string]struct{}) []storage.InputItem {
	log.Printf("Fetching a new item batch, offset %v..\n", highestTimestamp)
	// Fetch batch of items
	postgresItems, err := mig.ListFromPostgres(highestTimestamp)
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

		jsonItems = append(jsonItems, mig.ToJsonItem(item))
	}

	// Convert to input format (and mutates mysql db, inserting records etc)
	itemDb := storage.ItemDb{}
	inputItems, err := itemDb.ToInputItems(jsonItems)
	if err != nil {
		log.Fatalf("Error converting items to InputItem, %v", err)
	}

	log.Println("Finished fetching item batch..")
	return inputItems
}


func main() {
	mysql := config.GetDatabaseInstance()

	// 1. Get max timestamp from mysql
	var highestTimestamp int64
	row := mysql.Table("item").Select("max(ts)", "", 0).Row()
	row.Scan(&highestTimestamp)

	log.Printf("Migrating users..")
	migrateUsers()
	log.Printf("Users migrated..")



	log.Printf("Migrating items..")
	var hasMoreItems = true
	var lastInsertedItems = map[string]struct{}{}
	itemDb := storage.ItemDb{}
	for hasMoreItems {
		items := getItemBatch(highestTimestamp, lastInsertedItems)

		// Insert to mysql
		var currentlyInsertedItems = map[string]struct{}{}
		for _, item := range items {

			if err := itemDb.Insert(item.UserId, item); err != nil {
				log.Fatalf("Error inserting items to mysql, %v", err)
			}

			if item.Ts > highestTimestamp {
				highestTimestamp = item.Ts
			}

			currentlyInsertedItems[item.Id] = struct{}{}
		}
		lastInsertedItems = currentlyInsertedItems
		hasMoreItems = len(lastInsertedItems) > 0
	}
	log.Printf("Finished migrating items")



	log.Printf("Migrating item deletions")
	if err := mig.ResetItemDeletionInMysql(); err != nil {
		log.Fatalf("Error deleting items in mysql, %v", err)
	}

	deletedItems, err := mig.ListDeletedItemsFromPostgres()
	if err != nil {
		log.Fatalf("Error fetching deleted items, %v", err)
	}

	for _, item := range deletedItems {
		user := mig.GetUserByEmail(item.UserId)
		itemDb.Delete(user.UserId, item.Id, item.Ts)
	}


	log.Printf("Finished migrating item deletions")

	migrateCharacters()
}

func migrateCharacters() {
	log.Printf("Migrating characters..")

	characters, err := mig.ListCharactersFromPostgres()
	if err != nil {
		log.Fatalf("Error fetching characters, %v", err)
	}

	for _, entry := range characters {
		user := mig.GetUserByEmail(entry.Email)

		if err := mig.InsertCharactersToMysql(storage.CharacterEntry {
			UpdatedAt: entry.UpdatedAt,
			CreatedAt: entry.CreatedAt,
			Name: entry.Name,
			Filename: entry.Filename,
			UserId: user.UserId,
		}); err != nil {
			log.Fatalf("Error inserting character, %v", err)
		}
	}
	log.Printf("Characters migrated..")
}