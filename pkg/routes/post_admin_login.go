package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/phillip-england/thispage/pkg/validators"
	"github.com/phillip-england/vii/vii"
)

func PostAdminLogin(w http.ResponseWriter, r *http.Request) {
	validator := validators.FormAdminLogin{}
	data, err := validator.Validate(r)
	if err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if data.Username != adminUsername || data.Password != adminPassword {
		// In a real app, you might want to show the form again with an error
		vii.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Success - for now just redirect or say success
	// vii.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	fmt.Fprintln(w, "Login Successful")
}
