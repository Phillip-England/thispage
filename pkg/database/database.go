package database

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(projectPath string) error {
	dbPath := filepath.Join(projectPath, "data.db")
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	query := `
    CREATE TABLE IF NOT EXISTS session (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        key TEXT NOT NULL,
        token TEXT NOT NULL,
        expires_at DATETIME NOT NULL
    );
    `
	_, err = DB.Exec(query)
	return err
}
