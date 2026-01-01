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
	
    // Security: Only allow partials/ and templates/
    allowed := false
    for _, prefix := range []string{"partials", "templates"} {
        if strings.HasPrefix(destDir, prefix+string(os.PathSeparator)) || destDir == prefix {
            allowed = true
            break
        }
    }
     if !allowed {
        for _, prefix := range []string{"partials/", "templates/"} {
             if strings.HasPrefix(filepath.ToSlash(destDir), prefix) {
                allowed = true
                break
            }
        }
    }

	if !allowed {
		vii.WriteError(w, http.StatusForbidden, "Access denied: Cannot upload to restricted directory.")
		return
	}

	absDir := filepath.Join(projectPath, destDir)
	
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
