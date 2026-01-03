package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/phillip-england/thispage/pkg/database"
)

// LoginStatus represents the current rate limit status for an IP
type LoginStatus struct {
	IsBlacklisted  bool
	FailedAttempts int
	AttemptsLeft   int
}

// GetClientIP extracts the client IP address from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// IsBlacklisted checks if an IP address is in the blacklist
func IsBlacklisted(ip string) (bool, error) {
	var count int
	err := database.DB.QueryRow(
		"SELECT COUNT(*) FROM LOGIN_BLACKLIST WHERE ip_address = ?",
		ip,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRecentFailedAttempts returns the count of failed attempts within the time window
func GetRecentFailedAttempts(ip string) (int, error) {
	windowStart := time.Now().Add(-time.Duration(database.AttemptWindowSeconds) * time.Second)
	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM LOGIN_ATTEMPT
		WHERE ip_address = ? AND success = 0 AND attempted_at > ?
	`, ip, windowStart).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetLoginStatus returns the current rate limit status for an IP
func GetLoginStatus(ip string) (LoginStatus, error) {
	status := LoginStatus{}

	// Check blacklist
	blacklisted, err := IsBlacklisted(ip)
	if err != nil {
		return status, err
	}
	status.IsBlacklisted = blacklisted

	if blacklisted {
		status.AttemptsLeft = 0
		return status, nil
	}

	// Get failed attempts count
	failedCount, err := GetRecentFailedAttempts(ip)
	if err != nil {
		return status, err
	}
	status.FailedAttempts = failedCount
	status.AttemptsLeft = database.FailedAttemptThreshold - failedCount

	if status.AttemptsLeft < 0 {
		status.AttemptsLeft = 0
	}

	return status, nil
}

// evictOldestAttempt removes the oldest entry from LOGIN_ATTEMPT if at capacity
func evictOldestAttempt() error {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM LOGIN_ATTEMPT").Scan(&count)
	if err != nil {
		return err
	}

	if count >= database.MaxLoginAttempts {
		_, err = database.DB.Exec(`
			DELETE FROM LOGIN_ATTEMPT
			WHERE id = (SELECT id FROM LOGIN_ATTEMPT ORDER BY attempted_at ASC LIMIT 1)
		`)
		if err != nil {
			return err
		}
	}
	return nil
}

// evictOldestBlacklist removes the oldest entry from LOGIN_BLACKLIST if at capacity
func evictOldestBlacklist() error {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM LOGIN_BLACKLIST").Scan(&count)
	if err != nil {
		return err
	}

	if count >= database.MaxBlacklistEntries {
		_, err = database.DB.Exec(`
			DELETE FROM LOGIN_BLACKLIST
			WHERE id = (SELECT id FROM LOGIN_BLACKLIST ORDER BY blacklisted_at ASC LIMIT 1)
		`)
		if err != nil {
			return err
		}
	}
	return nil
}

// RecordAttempt records a login attempt and returns whether the IP should be blacklisted
func RecordAttempt(ip string, success bool) (shouldBlacklist bool, err error) {
	// First check blacklist before eviction
	blacklisted, err := IsBlacklisted(ip)
	if err != nil {
		return false, err
	}
	if blacklisted {
		return false, nil // Already blacklisted, don't record
	}

	// Evict oldest if at capacity
	if err := evictOldestAttempt(); err != nil {
		return false, err
	}

	// Record the attempt
	successInt := 0
	if success {
		successInt = 1
	}
	_, err = database.DB.Exec(`
		INSERT INTO LOGIN_ATTEMPT (ip_address, attempted_at, success)
		VALUES (?, ?, ?)
	`, ip, time.Now(), successInt)
	if err != nil {
		return false, err
	}

	// If successful login, no need to check for blacklist
	if success {
		return false, nil
	}

	// Check if we should blacklist this IP
	failedCount, err := GetRecentFailedAttempts(ip)
	if err != nil {
		return false, err
	}

	return failedCount >= database.FailedAttemptThreshold, nil
}

// AddToBlacklist adds an IP to the blacklist
func AddToBlacklist(ip string) error {
	// Evict oldest if at capacity
	if err := evictOldestBlacklist(); err != nil {
		return err
	}

	// Add to blacklist (ignore if already exists due to UNIQUE constraint)
	_, err := database.DB.Exec(`
		INSERT OR IGNORE INTO LOGIN_BLACKLIST (ip_address, blacklisted_at)
		VALUES (?, ?)
	`, ip, time.Now())
	return err
}

// ClearAttemptsForIP clears all login attempts for an IP (call on successful login)
func ClearAttemptsForIP(ip string) error {
	_, err := database.DB.Exec("DELETE FROM LOGIN_ATTEMPT WHERE ip_address = ?", ip)
	return err
}
