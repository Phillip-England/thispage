package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminDirCreate(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	parentDir := r.FormValue("parent_directory")
	dirname := r.FormValue("dirname")

	if parentDir == "" || dirname == "" {
		vii.WriteError(w, http.StatusBadRequest, "Parent directory and folder name are required")
		return
	}

	// Validate Parent Path
	parentDir = filepath.Clean(parentDir)
    parentDirSlash := filepath.ToSlash(parentDir)

    // Security: Only allow creating dirs inside allowed roots
    allowed := false
    if strings.HasPrefix(parentDirSlash, "templates") || strings.HasPrefix(parentDirSlash, "partials") || strings.HasPrefix(parentDirSlash, "static") {
        allowed = true
    }

	if !allowed {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted parent directory.")
		return
	}

    // Validate New Dirname (simple check)
    if strings.Contains(dirname, "/") || strings.Contains(dirname, "\\") || strings.Contains(dirname, ".") {
        vii.WriteError(w, http.StatusBadRequest, "Invalid folder name: Must be a simple name, not a path or file.")
		return
    }

	absParent := filepath.Join(projectPath, parentDir)
    absNewDir := filepath.Join(absParent, dirname)
    
    // Path traversal check
    if !strings.HasPrefix(absNewDir, projectPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Path traversal detected.")
        return
    }

    // Check if exists
    if _, err := os.Stat(absNewDir); err == nil {
         vii.WriteError(w, http.StatusConflict, "Directory already exists")
		 return
    }

	// Create directory
	err := os.MkdirAll(absNewDir, 0755)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error creating directory: "+err.Error())
		return
	}

	// Redirect to files list
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
