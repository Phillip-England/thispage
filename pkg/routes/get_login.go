package routes

import (
	"net/http"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/vii/vii"
)

func GetLogin(w http.ResponseWriter, r *http.Request) {
    if auth.IsAuthenticated(r) {
        http.Redirect(w, r, "/admin", http.StatusSeeOther)
        return
    }
	err := vii.Render(w, r, "admin_login.html", map[string]interface{}{})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
