package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminFileRename(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	oldRelPath := r.FormValue("old_path")
	newName := r.FormValue("new_name")

	if oldRelPath == "" || newName == "" {
		vii.WriteError(w, http.StatusBadRequest, "Old path and new name are required")
		return
	}

	// Validate Old Path
	oldRelPath = filepath.Clean(oldRelPath)
    if !isPathAllowed(oldRelPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted directory.")
		return
    }

    // Construct New Path
    // We assume newName is just the filename, not a full path.
    // If user provides a path separator in newName, we reject or handle it.
    // Let's strictly check for just a filename for safety/simplicity first.
    if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
        vii.WriteError(w, http.StatusBadRequest, "Invalid new name: Must be a filename, not a path.")
		return
    }

    parentDir := filepath.Dir(oldRelPath)
    newRelPath := filepath.Join(parentDir, newName)

    if !isPathAllowed(newRelPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Target directory restricted.")
		return
    }

	absOldPath := filepath.Join(projectPath, oldRelPath)
	absNewPath := filepath.Join(projectPath, newRelPath)

    // Check if destination already exists
    if _, err := os.Stat(absNewPath); err == nil {
         vii.WriteError(w, http.StatusConflict, "A file with that name already exists.")
		return
    }

	err := os.Rename(absOldPath, absNewPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error renaming file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}

// Helper to reuse the security logic
func isPathAllowed(relPath string) bool {
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
    return allowed
}
