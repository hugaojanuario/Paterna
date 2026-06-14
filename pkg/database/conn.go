package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewConnection() (*sql.DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("database: get home dir: %w", err)
	}

	dir := filepath.Join(home, ".paterna")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("database: create paterna dir: %w", err)
	}

	dbPath := filepath.Join(dir, "paterna.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return db, nil
}
