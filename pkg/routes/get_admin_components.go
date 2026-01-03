package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func GetAdminComponents(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	componentsDir := filepath.Join(projectPath, "components")
	files, err := os.ReadDir(componentsDir)
	if err != nil {
		if os.IsNotExist(err) {
			vii.WriteJSON(w, http.StatusOK, map[string][]string{"files": {}})
			return
		}
		vii.WriteError(w, http.StatusInternalServerError, "Error reading components: "+err.Error())
		return
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			fileNames = append(fileNames, file.Name())
		}
	}

	vii.WriteJSON(w, http.StatusOK, map[string][]string{"files": fileNames})
}
