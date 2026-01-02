package routes

import (
	"net/http"
	"os"

	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/thispage/pkg/forms"
	"github.com/phillip-england/vii/vii"
)

func PostLogin(w http.ResponseWriter, r *http.Request) {
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

    // Create Session
    if err := auth.CreateSession(w, r); err != nil {
        vii.WriteError(w, http.StatusInternalServerError, "Failed to create session: "+err.Error())
        return
    }

	// Success
	vii.Redirect(w, r, "/admin", http.StatusSeeOther)
}
