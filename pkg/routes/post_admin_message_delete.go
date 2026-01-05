package routes

import (
	"net/http"
	"strconv"

	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/vii/vii"
)

func PostAdminMessageDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	idStr := r.Form.Get("id")
	if idStr == "" {
		vii.WriteError(w, http.StatusBadRequest, "Message ID is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	_, err = database.DB.Exec("DELETE FROM ADMIN_MESSAGE WHERE id = ?", id)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to delete message: "+err.Error())
		return
	}

	vii.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}
