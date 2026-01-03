package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/phillip-england/thispage/pkg/credentials"
	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func GenerateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Cleanup() error {
	_, err := database.DB.Exec("DELETE FROM session WHERE expires_at < ?", time.Now())
	return err
}

func CreateSession(w http.ResponseWriter, r *http.Request) error {
	// Cleanup old sessions first
	if err := Cleanup(); err != nil {
		// Log error but continue?
	}

	key, err := GenerateKey()
	if err != nil {
		return err
	}

	token, err := sessionTokenFromRequest(r)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = database.DB.Exec("INSERT INTO session (key, token, expires_at) VALUES (?, ?, ?)", key, token, expiresAt)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    key,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiresAt,
	})

	return nil
}

func IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		return false
	}

	key := cookie.Value
	tokenEnv, err := sessionTokenFromRequest(r)
	if err != nil {
		return false
	}

	var tokenDb string
	var expiresAt time.Time

	row := database.DB.QueryRow("SELECT token, expires_at FROM session WHERE key = ?", key)
	err = row.Scan(&tokenDb, &expiresAt)
	if err != nil {
		return false
	}

	if time.Now().After(expiresAt) {
		return false
	}

	if tokenDb != tokenEnv {
		return false
	}

	return true
}

// IsAuthenticatedAndRefresh checks authentication and extends the session
func IsAuthenticatedAndRefresh(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		return false
	}

	key := cookie.Value
	tokenEnv, err := sessionTokenFromRequest(r)
	if err != nil {
		return false
	}

	var tokenDb string
	var expiresAt time.Time

	row := database.DB.QueryRow("SELECT token, expires_at FROM session WHERE key = ?", key)
	err = row.Scan(&tokenDb, &expiresAt)
	if err != nil {
		return false
	}

	if time.Now().After(expiresAt) {
		return false
	}

	if tokenDb != tokenEnv {
		return false
	}

	// Extend session by 15 more minutes
	newExpiry := time.Now().Add(15 * time.Minute)
	database.DB.Exec("UPDATE session SET expires_at = ? WHERE key = ?", newExpiry, key)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    key,
		Path:     "/",
		HttpOnly: true,
		Expires:  newExpiry,
	})

	return true
}

func DeleteSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		return nil // No session to delete
	}

	key := cookie.Value
	_, err = database.DB.Exec("DELETE FROM session WHERE key = ?", key)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
		MaxAge:   -1,
	})

	return nil
}

func sessionTokenFromRequest(r *http.Request) (string, error) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok || projectPath == "" {
		return "", fmt.Errorf("project path not found in context")
	}

	_, _, token, err := credentials.LoadWithToken(projectPath)
	if err != nil {
		return "", err
	}
	return token, nil
}
