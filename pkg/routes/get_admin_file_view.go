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
	// Normalize to forward slashes for checking
	slashPath := filepath.ToSlash(relPath)

	// Security: Check allowed directories and extensions
	allowed := false
	isImage := false
	
	if strings.HasPrefix(slashPath, "templates/static/") {
		ext := strings.ToLower(filepath.Ext(slashPath))
		switch ext {
		case ".css", ".js":
			allowed = true
		case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
			allowed = true
			isImage = true
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
		vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted file or directory.")
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

    data := map[string]interface{}{
		"ProjectPath": projectPath,
		"FilePath":    relPath,
	}

	if isImage {
        data["IsImage"] = true
        // Map templates/static/foo.png -> /static/foo.png
        data["Src"] = "/static/" + strings.TrimPrefix(slashPath, "templates/static/")
	} else {
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
        data["Content"] = string(content)
        data["IsEditable"] = true

        // Calculate LiveURL for templates
        if strings.HasPrefix(slashPath, "templates/") && !strings.HasPrefix(slashPath, "templates/static/") && strings.HasSuffix(slashPath, ".html") {
            // templates/index.html -> /index.html
            // templates/posts/view.html -> /posts/view.html
            subPath := strings.TrimPrefix(slashPath, "templates/")
            data["LiveURL"] = "/" + subPath
        }
    }

	err := vii.Render(w, r, "admin_file_view.html", data)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}
