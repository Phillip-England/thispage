package routes

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminFileUpload(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	// 10 MB limit
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Error retrieving file: "+err.Error())
		return
	}
	defer file.Close()

	destDir := r.FormValue("directory")
	if destDir == "" {
		destDir = "static" // Default to static
	}

	// Clean and Validate Path
	destDir = filepath.Clean(destDir)
    
    // Construct the full relative path of the file to be saved
    relPath := filepath.Join(destDir, handler.Filename)
    slashPath := filepath.ToSlash(relPath)

    // Security: Validate file type based on directory
    allowed := false
    
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
	} else if strings.HasPrefix(slashPath, "partials/") {
		if strings.HasSuffix(slashPath, ".html") {
			allowed = true
		}
	}

	if !allowed {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Invalid file type for this directory.")
		return
	}

	absDir := filepath.Join(projectPath, destDir)
    // Double check absDir is inside projectPath
    if !strings.HasPrefix(absDir, projectPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Path traversal detected.")
        return
    }
	
	// Ensure the directory exists
	if err := os.MkdirAll(absDir, 0755); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error creating directory: "+err.Error())
		return
	}

	dstPath := filepath.Join(absDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error creating file: "+err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error saving file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}
