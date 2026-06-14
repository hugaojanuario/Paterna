package database

import (
	"database/sql"
	"fmt"

	"github.com/hugaojanuario/Paterna/internal/shared/config"
	_ "github.com/lib/pq"
)

func NewConnection(config config.Config) (*sql.DB, error) {
	strgDBConfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHOST, config.DBPORT, config.DBUSER, config.DBPASSWORD, config.DBNAME, config.DBSSLMODE)

	db, err := sql.Open("postgres", strgDBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
