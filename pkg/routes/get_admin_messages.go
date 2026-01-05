package routes

import (
	"net/http"
	"time"

	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/vii/vii"
)

type AdminMessage struct {
	ID        int
	IPAddress string
	Name      string
	Email     string
	Message   string
	CreatedAt time.Time
}

func GetAdminMessages(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(`
		SELECT id, ip_address, name, email, message, created_at
		FROM ADMIN_MESSAGE
		ORDER BY created_at DESC
	`)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to fetch messages: "+err.Error())
		return
	}
	defer rows.Close()

	var messages []AdminMessage
	for rows.Next() {
		var msg AdminMessage
		if err := rows.Scan(&msg.ID, &msg.IPAddress, &msg.Name, &msg.Email, &msg.Message, &msg.CreatedAt); err != nil {
			vii.WriteError(w, http.StatusInternalServerError, "Failed to scan message: "+err.Error())
			return
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error reading messages: "+err.Error())
		return
	}

	var totalCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM ADMIN_MESSAGE").Scan(&totalCount)

	vii.Render(w, r, "admin_messages.html", map[string]interface{}{
		"Messages":   messages,
		"TotalCount": totalCount,
		"MaxCount":   database.MaxAdminMessages,
	})
}
