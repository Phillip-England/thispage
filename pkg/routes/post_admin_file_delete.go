package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminFileDelete(w http.ResponseWriter, r *http.Request) {
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
	if relPath == "" {
		vii.WriteError(w, http.StatusBadRequest, "Path is required")
		return
	}

	// Security Check
	relPath = filepath.Clean(relPath)
	
    // Security: Only allow partials/ and templates/
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

    // Verify it exists before trying to delete (optional but good for error messaging)
    if _, err := os.Stat(absPath); os.IsNotExist(err) {
        vii.WriteError(w, http.StatusNotFound, "File or directory not found")
        return
    }

	// Delete recursively
	err := os.RemoveAll(absPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error deleting file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}
