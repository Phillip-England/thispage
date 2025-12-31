package routes

import (
	"net/http"

	"github.com/phillip-england/vii/vii"
)

func GetAdminDashboard(w http.ResponseWriter, r *http.Request) {
	projectPath, _ := vii.GetContext("PROJECT_PATH", r).(string)
	
	// In a real app, check for authentication cookie here
	err := vii.Render(w, r, "admin_dashboard.html", map[string]interface{}{
		"ProjectPath": projectPath,
	})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
