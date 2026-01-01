package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func GetAdminFileView(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	// Get the relative path from the query parameter
	relPath := r.URL.Query().Get("path")
	if relPath == "" {
		vii.WriteError(w, http.StatusBadRequest, "Path parameter is required")
		return
	}

	// Construct the absolute path
	// clean the path to resolve .. etc
	relPath = filepath.Clean(relPath)
	
    // Security: Only allow partials/ and templates/
    allowed := false
    for _, prefix := range []string{"partials", "templates"} {
        // We use string concatenation with separator to ensure we match directory boundaries
        // or exact match if we allowed top-level files (which we don't effectively, as we only iterate those dirs)
        // clean path "partials/file.html" starts with "partials"
        if strings.HasPrefix(relPath, prefix+string(os.PathSeparator)) || relPath == prefix {
            allowed = true
            break
        }
    }
    
    // Check for forward slash style too if running on windows but using slash urls
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

	// Prevent directory traversal: ensure the path doesn't start with .. or /../ or is just ..
    // Since we join with projectPath, we just need to ensure the resulting path starts with projectPath
	absPath := filepath.Join(projectPath, relPath)

	// Verify the path is still within the project directory
	if !strings.HasPrefix(absPath, projectPath) {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Path outside project directory")
		return
	}

	// Read the file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			vii.WriteError(w, http.StatusNotFound, "File not found")
			return
		}
		vii.WriteError(w, http.StatusInternalServerError, "Error reading file: "+err.Error())
		return
	}

	err = vii.Render(w, r, "admin_file_view.html", map[string]interface{}{
		"ProjectPath": projectPath,
		"FilePath":    relPath,
		"Content":     string(content),
	})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
