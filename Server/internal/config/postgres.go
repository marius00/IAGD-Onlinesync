package config

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"os"
)

var postgres *gorm.DB

func GetPostgresInstance() *gorm.DB {
	if postgres == nil {
		log.Printf("Opening database connection to %s, db %s..\n", os.Getenv("PG_DATABASE_HOST"), os.Getenv("PG_DATABASE_NAME"))

		connectionString := fmt.Sprintf(
			"host=%s user=%s dbname=%s password=%s sslmode=disable connect_timeout=5",
			os.Getenv("PG_DATABASE_HOST"),
			os.Getenv("PG_DATABASE_USER"),
			os.Getenv("PG_DATABASE_NAME"),
			os.Getenv("PG_DATABASE_PASSWORD"),
		)

		newDb, err := gorm.Open("postgres", connectionString)

		if err != nil {
			panic(err)
		}

		postgres = newDb
		postgres.SingularTable(true)
		postgres.LogMode(true)
	}

	return postgres
}
