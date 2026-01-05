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

	// Support both single "id" and multiple "ids" parameters
	ids := r.Form["ids"]

	if len(ids) == 0 {
		vii.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
		return
	}

	// Delete each message
	for _, idStr := range ids {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		database.DB.Exec("DELETE FROM ADMIN_MESSAGE WHERE id = ?", id)
	}

	vii.Redirect(w, r, "/admin/messages", http.StatusSeeOther)
}
