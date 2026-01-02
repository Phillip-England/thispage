package routes

import (
	"net/http"
	"os"

	"github.com/phillip-england/thispage/pkg/forms"
	"github.com/phillip-england/vii/vii"
)

func PostAdmin(w http.ResponseWriter, r *http.Request) {
	validator := forms.FormAdminLogin{}
	data, err := validator.Validate(r)
	if err != nil {
		vii.Render(w, r, "admin_login.html", map[string]interface{}{"Error": "Invalid form data"})
		return
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if data.Username != adminUsername || data.Password != adminPassword {
		vii.Render(w, r, "admin_login.html", map[string]interface{}{"Error": "Invalid credentials"})
		return
	}

	// Success
	vii.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}
