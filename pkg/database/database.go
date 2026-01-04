package database

import (
	"database/sql"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// MaxLoginAttempts is the maximum number of entries in the LOGIN_ATTEMPT table
const MaxLoginAttempts = 1000

// MaxBlacklistEntries is the maximum number of entries in the LOGIN_BLACKLIST table
const MaxBlacklistEntries = 1000

// FailedAttemptThreshold is the number of failed attempts before blacklisting
const FailedAttemptThreshold = 5

// AttemptWindowSeconds is the time window (in seconds) for counting consecutive failures
const AttemptWindowSeconds = 60

func Init(projectPath string) error {
	dbPath := filepath.Join(projectPath, "data.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath)
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

    CREATE TABLE IF NOT EXISTS LOGIN_ATTEMPT (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address TEXT NOT NULL,
        attempted_at DATETIME NOT NULL,
        success INTEGER NOT NULL DEFAULT 0
    );

    CREATE INDEX IF NOT EXISTS idx_login_attempt_ip ON LOGIN_ATTEMPT(ip_address);
    CREATE INDEX IF NOT EXISTS idx_login_attempt_time ON LOGIN_ATTEMPT(attempted_at);

    CREATE TABLE IF NOT EXISTS LOGIN_BLACKLIST (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address TEXT NOT NULL UNIQUE,
        blacklisted_at DATETIME NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_blacklist_ip ON LOGIN_BLACKLIST(ip_address);
    `
	_, err = DB.Exec(query)
	return err
}
