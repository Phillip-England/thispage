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

// MaxAdminMessages is the maximum number of messages in the ADMIN_MESSAGE table (queue)
const MaxAdminMessages = 100

// MaxMessagesPerIPPerDay is the daily message limit per IP address
const MaxMessagesPerIPPerDay = 3

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

    CREATE TABLE IF NOT EXISTS ADMIN_MESSAGE (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address TEXT NOT NULL,
        name TEXT NOT NULL,
        email TEXT NOT NULL,
        message TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );

    CREATE INDEX IF NOT EXISTS idx_admin_message_ip ON ADMIN_MESSAGE(ip_address);
    CREATE INDEX IF NOT EXISTS idx_admin_message_created ON ADMIN_MESSAGE(created_at);
    `
	_, err = DB.Exec(query)
	return err
}
