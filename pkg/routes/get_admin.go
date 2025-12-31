package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

func GetAdmin(w http.ResponseWriter, r *http.Request) {
	err := vii.Render(w, r, "admin_login.html", map[string]interface{}{})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
