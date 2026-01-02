package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func GetAdminPartials(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	partialsDir := filepath.Join(projectPath, "partials")
	files, err := os.ReadDir(partialsDir)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error reading partials: "+err.Error())
		return
	}

	var filenames []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			filenames = append(filenames, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": filenames,
	})
}
