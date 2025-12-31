package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

func GetAdminLogout(w http.ResponseWriter, r *http.Request) {
	// In a real app, clear the session cookie here
	vii.Redirect(w, r, "/admin", http.StatusSeeOther)
}
