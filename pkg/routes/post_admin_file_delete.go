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
    slashPath := filepath.ToSlash(relPath)
	
    // Prevent deleting root directories
    if slashPath == "templates" || slashPath == "components" || slashPath == "static" || slashPath == "layouts" {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Cannot delete root directories.")
        return
    }

	absPath := filepath.Join(projectPath, relPath)
    
    // Path traversal check
    if !strings.HasPrefix(absPath, projectPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Path traversal detected.")
        return
    }

    // Verify it exists
    info, err := os.Stat(absPath)
    if err != nil {
        if os.IsNotExist(err) {
            vii.WriteError(w, http.StatusNotFound, "File or directory not found")
            return
        }
         vii.WriteError(w, http.StatusInternalServerError, "Error accessing path: "+err.Error())
         return
    }

    // Security: Validate file type based on directory
    allowed := false
    
    // Check prefix first
    if strings.HasPrefix(slashPath, "templates/") || strings.HasPrefix(slashPath, "components/") || strings.HasPrefix(slashPath, "static/") || strings.HasPrefix(slashPath, "layouts/") {
        if info.IsDir() {
            // Allow deleting subdirectories
            allowed = true
        } else {
             // File checks
             if strings.HasPrefix(slashPath, "static/") {
                ext := strings.ToLower(filepath.Ext(slashPath))
                switch ext {
                case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
                    allowed = true
                }
            } else if strings.HasPrefix(slashPath, "templates/") {
                if strings.HasSuffix(slashPath, ".html") {
                    allowed = true
                }
            } else if strings.HasPrefix(slashPath, "components/") {
                if strings.HasSuffix(slashPath, ".html") {
                    allowed = true
                }
            } else if strings.HasPrefix(slashPath, "layouts/") {
                if strings.HasSuffix(slashPath, ".html") {
                    allowed = true
                }
            }
        }
    }

	if !allowed {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Invalid file type or directory.")
		return
	}

	// Delete recursively
	err = os.RemoveAll(absPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error deleting file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
