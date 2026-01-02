package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminInsertPartial(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	targetFile := r.FormValue("target_file")
	partialName := r.FormValue("partial_name")

	if targetFile == "" || partialName == "" {
		vii.WriteError(w, http.StatusBadRequest, "Target file and partial name are required")
		return
	}

    // Security Check
    targetFile = filepath.Clean(targetFile)
    if !strings.HasPrefix(targetFile, "templates") {
        vii.WriteError(w, http.StatusForbidden, "Invalid target file")
        return
    }

	absPath := filepath.Join(projectPath, targetFile)
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error reading file: "+err.Error())
		return
	}
    
    content := string(contentBytes)
    insertion := fmt.Sprintf("\n    {{ include ./%s }}\n", partialName)

    // Strategy: Try to insert before </main>
    if strings.Contains(content, "</main>") {
        content = strings.Replace(content, "</main>", insertion+"    </main>", 1)
    } else if strings.Contains(content, "</body>") {
         content = strings.Replace(content, "</body>", insertion+"</body>", 1)
    } else {
        content += insertion
    }

	err = os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error saving file: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
