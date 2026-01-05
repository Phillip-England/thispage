package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/thispage/pkg/forms"
	"github.com/phillip-england/thispage/pkg/ratelimit"
	"github.com/phillip-england/vii/vii"
)

func PostContact(w http.ResponseWriter, r *http.Request) {
	clientIP := ratelimit.GetClientIP(r)

	// Check daily message limit for this IP
	messagesRemaining, err := getMessagesRemainingToday(clientIP)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to check message limit: "+err.Error())
		return
	}

	if messagesRemaining <= 0 {
		vii.Render(w, r, "admin_contact.html", map[string]interface{}{
			"Error":     "You have reached your daily message limit. Please try again tomorrow.",
			"IsBlocked": true,
		})
		return
	}

	// Validate form
	validator := forms.FormContact{}
	data, err := validator.Validate(r)
	if err != nil {
		vii.Render(w, r, "admin_contact.html", map[string]interface{}{
			"Error":             err.Error(),
			"MessagesRemaining": messagesRemaining,
		})
		return
	}

	// Evict oldest message if at capacity (queue behavior)
	if err := evictOldestMessage(); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to manage message queue: "+err.Error())
		return
	}

	// Insert the message
	_, err = database.DB.Exec(`
		INSERT INTO ADMIN_MESSAGE (ip_address, name, email, message, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, clientIP, data.Name, data.Email, data.Message, time.Now())
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to save message: "+err.Error())
		return
	}

	vii.Render(w, r, "admin_contact.html", map[string]interface{}{
		"Success":           true,
		"MessagesRemaining": messagesRemaining - 1,
	})
}

// getMessagesRemainingToday returns how many messages an IP can still send today
func getMessagesRemainingToday(ip string) (int, error) {
	// Get start of today (midnight UTC)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM ADMIN_MESSAGE
		WHERE ip_address = ? AND created_at >= ?
	`, ip, startOfDay).Scan(&count)
	if err != nil {
		return 0, err
	}

	remaining := database.MaxMessagesPerIPPerDay - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// evictOldestMessage removes the oldest message if at capacity
func evictOldestMessage() error {
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM ADMIN_MESSAGE").Scan(&count)
	if err != nil {
		return err
	}

	if count >= database.MaxAdminMessages {
		_, err = database.DB.Exec(`
			DELETE FROM ADMIN_MESSAGE
			WHERE id = (SELECT id FROM ADMIN_MESSAGE ORDER BY created_at ASC LIMIT 1)
		`)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMessageStatus returns the message limit status for display purposes
func GetMessageStatus(r *http.Request) (remaining int, err error) {
	clientIP := ratelimit.GetClientIP(r)
	return getMessagesRemainingToday(clientIP)
}

// Ensure database constants are accessible
var _ = fmt.Sprint(database.MaxMessagesPerIPPerDay)
