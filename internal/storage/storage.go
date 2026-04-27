package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(connStr string) (*Storage, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	queryFromSites := "CREATE TABLE IF NOT EXISTS sites (id SERIAL PRIMARY KEY, url TEXT NOT NULL, last_checked_at TIMESTAMP NOT NULL, interval_seconds INT NOT NULL DEFAULT 60, is_DLQ BOOL DEFAULT false);"
	if _, err := db.Exec(queryFromSites); err != nil {
		return nil, fmt.Errorf("error create table: %v\n", err)
	}

	queryFromCheckTable := "CREATE TABLE IF NOT EXISTS checkStatus (id SERIAL PRIMARY KEY, url TEXT NOT NULL, status INT, error TEXT, requestAt TIMESTAMP, responseAt TIMESTAMP);"
	if _, err := db.Exec(queryFromCheckTable); err != nil {
		return nil, fmt.Errorf("error create check table: %v\n", err)
	}

	return &Storage{
		db: db,
	}, nil
}
