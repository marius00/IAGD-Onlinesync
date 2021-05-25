package main

import (
	"fmt"
	"github.com/marmyr/iagdbackup/internal/storage"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
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

func GetDatabaseInstance() *gorm.DB {
	if db == nil {
		log.Printf("Opening database connection to %s, db %s..\n", os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_NAME"))

		dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASSWORD"),
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_NAME"),
		)

		newDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			Logger: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             time.Second,
					LogLevel:                  logger.Info,
					IgnoreRecordNotFoundError: true,
					Colorful:                  false,
				},
			),
		})

		if err != nil {
			log.Fatal(err)
		} else {
			db = newDb
		}
	}

	return db
}

func main() {
	DB := GetDatabaseInstance()
	var items []storage.OutputItem
	result := DB.Where("userid = ? AND ts > ?", "", 0).Order("ts asc").Limit(100).Find(&items)
	fmt.Printf("%v, %v\n", items, result.Error)

	itemDb := storage.ItemDb{}
	ref, err := itemDb.GetRecordReferences([]storage.JsonItem{
		{BaseRecord: "records/0_rot/items/armor/npc/madawc/npc_madawc_legs01.dbr",},
		{BaseRecord: "records/0_rot/items/lootaffixes/prefix/classama/skill_08a.dbr",},
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", ref)

}

func List(user string, lastTimestamp int64) ([]storage.OutputItem, error) {
	DB := GetDatabaseInstance()

	var items []storage.OutputItem
	result := DB.Where("userid = ? AND ts > ?", user, lastTimestamp).Order("ts asc").Limit(100).Find(&items)

	return items, result.Error
}
