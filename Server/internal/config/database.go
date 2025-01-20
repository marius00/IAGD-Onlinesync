package config

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

var sqlxDb *sqlx.DB

func GetDatabaseInstance() *sqlx.DB {
	if sqlxDb == nil {
		log.Printf("Opening database connection to %s, db %s..\n", os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_NAME"))

		if os.Getenv("DATABASE_USER") == "" {
			log.Fatal().Msgf("DATABASE_USER is not set")
		}
		if os.Getenv("DATABASE_PASSWORD") == "" {
			log.Fatal().Msgf("DATABASE_PASSWORD is not set")
		}
		if os.Getenv("DATABASE_HOST") == "" {
			log.Fatal().Msgf("DATABASE_HOST is not set")
		}
		if os.Getenv("DATABASE_NAME") == "" {
			log.Fatal().Msgf("DATABASE_NAME is not set")
		}

		datasource := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s",
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_PASSWORD"),
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_NAME"),
		)

		newDb, err := sqlx.Open("mysql", datasource)
		if err != nil {
			log.Fatal().Msgf("Error opening mysql connection, %v", err)
		}

		newDb.SetConnMaxLifetime(time.Minute * 3)
		newDb.SetMaxOpenConns(300)
		newDb.SetMaxIdleConns(10)

		err = newDb.Ping()
		if err != nil {
			log.Warn().Msgf("Error pinging DB on %s", strings.Replace(datasource, os.Getenv("DATABASE_PASSWORD"), "REDACTED", -1))
			log.Fatal().Msgf("Could not ping db, %v", err)
		}

		sqlxDb = newDb
	}

	return sqlxDb
}
