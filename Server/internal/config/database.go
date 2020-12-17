package config

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB
// TODO: Just use github.com/jackc/pgx instead
// https://github.com/jackc/pgx
// https://medium.com/avitotech/how-to-work-with-postgres-in-go-bad2dabd13e4
func GetDatabaseInstance() *gorm.DB {
	if db == nil {
		log.Printf("Opening database connection to %s, db %s..\n", os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_NAME"))

		connectionString := fmt.Sprintf(
			"host=%s user=%s dbname=%s password=%s sslmode=disable connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_PASSWORD"),
		)
		/*conn, err := pgx.Connect(context.Background(), connectionString)
		if err != nil {
			panic(err)
		}*/

		newDb, err := gorm.Open("postgres", connectionString)

		if err != nil {
			panic(err)
		}

		db = newDb
		db.SingularTable(true)
		db.LogMode(true)
	}

	return db
}
