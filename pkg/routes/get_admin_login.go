package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

func GetAdminLogin(w http.ResponseWriter, r *http.Request) {
	err := vii.Render(w, r, "admin_login.html", nil)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
