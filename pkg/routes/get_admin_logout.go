package routes

import (
	"net/http"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/vii/vii"
)

func GetAdminLogout(w http.ResponseWriter, r *http.Request) {
    _ = auth.DeleteSession(w, r)
	vii.Redirect(w, r, "/login", http.StatusSeeOther)
}
