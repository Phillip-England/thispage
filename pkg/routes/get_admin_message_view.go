package routes

import (
	"net/http"
	"strconv"

	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/vii/vii"
)

func GetAdminMessageView(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		vii.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var msg AdminMessage
	err = database.DB.QueryRow(`
		SELECT id, ip_address, name, email, message, created_at
		FROM ADMIN_MESSAGE
		WHERE id = ?
	`, id).Scan(&msg.ID, &msg.IPAddress, &msg.Name, &msg.Email, &msg.Message, &msg.CreatedAt)

	if err != nil {
		vii.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
		return
	}

	vii.Render(w, r, "admin_message_view.html", map[string]interface{}{
		"Message": msg,
	})
}
