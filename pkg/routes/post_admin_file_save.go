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

	// Security Check (same as View)
	relPath = filepath.Clean(relPath)

    // Security: Only allow partials/, templates/, and static/
    allowed := false
    for _, prefix := range []string{"partials", "templates", "static"} {
        if strings.HasPrefix(relPath, prefix+string(os.PathSeparator)) || relPath == prefix {
            allowed = true
            break
        }
    }
     if !allowed {
        for _, prefix := range []string{"partials/", "templates/", "static/"} {
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

	// Redirect back to the view page (or the files list, but staying on the view confirms the save and lets them keep editing)
	// We can add a success message or query param if we had a flash message system, but for now a simple redirect is fine.
	http.Redirect(w, r, "/admin/files/view?path="+relPath, http.StatusSeeOther)
}
