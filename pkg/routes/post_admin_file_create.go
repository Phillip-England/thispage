package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminFileCreate(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	destDir := r.FormValue("directory")
    filename := r.FormValue("filename")

	if destDir == "" || filename == "" {
		vii.WriteError(w, http.StatusBadRequest, "Directory and filename are required")
		return
	}

	// Validate Path
    // Combine dir and filename
    relPath := filepath.Join(destDir, filename)
	relPath = filepath.Clean(relPath)
	
    // Security Check
    allowed := false
    for _, prefix := range []string{"partials", "templates"} {
        if strings.HasPrefix(relPath, prefix+string(os.PathSeparator)) || relPath == prefix {
            allowed = true
            break
        }
    }
     if !allowed {
        for _, prefix := range []string{"partials/", "templates/"} {
             if strings.HasPrefix(filepath.ToSlash(relPath), prefix) {
                allowed = true
                break
            }
        }
    }

	if !allowed {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted directory.")
		return
	}

	absPath := filepath.Join(projectPath, relPath)
    absDir := filepath.Dir(absPath)

    // Ensure directory exists (if they typed a new folder path in filename e.g. "subdir/file.txt")
	if err := os.MkdirAll(absDir, 0755); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error creating directory: "+err.Error())
		return
	}

    // Check if file exists
    if _, err := os.Stat(absPath); err == nil {
         vii.WriteError(w, http.StatusConflict, "File already exists")
		 return
    }

    initialContent := r.FormValue("initial_content")

	// Create file with initial content
	err := os.WriteFile(absPath, []byte(initialContent), 0644)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error creating file: "+err.Error())
		return
	}

	// Redirect to files list
	http.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}
