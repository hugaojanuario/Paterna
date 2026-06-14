package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/hugaojanuario/Paterna/pkg/database"
)

var db *sql.DB

const schema = `
  CREATE TABLE IF NOT EXISTS users (
      id            INTEGER PRIMARY KEY AUTOINCREMENT,
      email         TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );`

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

func Init() error {
	conn, err := database.NewConnection()
	if err != nil {
		return fmt.Errorf("repository: init connection: %w", err)
	}
	if _, err := conn.Exec(schema); err != nil {
		return fmt.Errorf("repository: create schema: %w", err)
	}
	db = conn
	return nil
}

func Create(email, password string) error {
	if db == nil {
		return errors.New("repository: not initialized, call Init() first")
	}

	_, err := db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, string(password),
	)
	if err != nil {
		return fmt.Errorf("repository: insert user: %w", err)
	}
	return nil
}

func Delete(email string) error {
	if db == nil {
		return errors.New("repository: not initialized, call Init() first")
	}

	res, err := db.Exec(`DELETE FROM users WHERE email = ?`, email)
	if err != nil {
		return fmt.Errorf("repository: delete user: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("repository: user not found")
	}
	return nil
}

func GetByEmail(email string) (*User, error) {
	if db == nil {
		return nil, errors.New("repository: not initialized, call Init() first")
	}

	var u User
	err := db.QueryRow(
		`SELECT id, email, password_hash FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("repository: user not found")
		}
		return nil, fmt.Errorf("repository: query user: %w", err)
	}
	return &u, nil
}
