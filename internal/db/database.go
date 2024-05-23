package db

import (
	"database/sql"

	"github.com/Kasama/kasama-twitch-integrations/internal/logger"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

const localDatabase = "run/database.db"

var db *Database

func GetDatabase() (*Database, error) {
	if db == nil {
		d, err := sql.Open("sqlite3", localDatabase)
		if err != nil {
			logger.Errorf("Failed to open database: %s", err.Error())
			return nil, err
		}
		db = &Database{
			DB: d,
		}
	}
	return db, nil
}
