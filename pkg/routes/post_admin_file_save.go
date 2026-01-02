package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminFileSave(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	relPath := r.FormValue("path")
	content := r.FormValue("content")

	if relPath == "" {
		vii.WriteError(w, http.StatusBadRequest, "Path is required")
		return
	}

	// Security Check
	relPath = filepath.Clean(relPath)
	slashPath := filepath.ToSlash(relPath)

    // Security: Validate file type based on directory
    allowed := false
    
    if strings.HasPrefix(slashPath, "templates/static/") {
		ext := strings.ToLower(filepath.Ext(slashPath))
		switch ext {
		case ".css", ".js":
            // Only text-based assets are editable
			allowed = true
		}
	} else if strings.HasPrefix(slashPath, "templates/") {
		if strings.HasSuffix(slashPath, ".html") {
			allowed = true
		}
	} else if strings.HasPrefix(slashPath, "partials/") {
		if strings.HasSuffix(slashPath, ".html") {
			allowed = true
		}
	}

    if !allowed {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Invalid file type or directory for text editing.")
        return
    }

	absPath := filepath.Join(projectPath, relPath)

	if !strings.HasPrefix(absPath, projectPath) {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Path outside project directory")
		return
	}

	// Write the file
	// 0644 is a good default permission for files (rw-r--r--)
    // We are overwriting the file.
	err := os.WriteFile(absPath, []byte(content), 0644)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error saving file: "+err.Error())
		return
	}

	// Redirect back to the view page
	http.Redirect(w, r, "/admin/files/view?path="+relPath, http.StatusSeeOther)
}
