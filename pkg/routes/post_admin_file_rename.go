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

	oldRelPath = filepath.Clean(oldRelPath)
    slashOldPath := filepath.ToSlash(oldRelPath)
    
    // Prevent renaming root dirs
    if slashOldPath == "templates" || slashOldPath == "partials" || slashOldPath == "templates/static" {
         vii.WriteError(w, http.StatusForbidden, "Access denied: Cannot rename root directories.")
		 return
    }

	absOldPath := filepath.Join(projectPath, oldRelPath)
    info, err := os.Stat(absOldPath)
    if err != nil {
         if os.IsNotExist(err) {
            vii.WriteError(w, http.StatusNotFound, "File not found")
            return
        }
		vii.WriteError(w, http.StatusInternalServerError, "Error accessing file: "+err.Error())
		return
    }
    isDir := info.IsDir()

    // Validate Old Path
    if !isPathAllowed(oldRelPath, isDir) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted source.")
		return
    }

    // Construct New Path
    if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
        vii.WriteError(w, http.StatusBadRequest, "Invalid new name: Must be a filename, not a path.")
		return
    }

    parentDir := filepath.Dir(oldRelPath)
    newRelPath := filepath.Join(parentDir, newName)

    // Validate New Path
    if !isPathAllowed(newRelPath, isDir) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted destination (invalid type or directory).")
		return
    }

	absNewPath := filepath.Join(projectPath, newRelPath)

    // Check if destination already exists
    if _, err := os.Stat(absNewPath); err == nil {
         vii.WriteError(w, http.StatusConflict, "A file with that name already exists.")
		return
    }

	err = os.Rename(absOldPath, absNewPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Error renaming file: "+err.Error())
		return
	}

	http.Redirect(w, r, "/admin/files", http.StatusSeeOther)
}

func isPathAllowed(relPath string, isDir bool) bool {
    slashPath := filepath.ToSlash(relPath)

    // If it's a directory, we mainly check if it's within permitted roots
    if isDir {
        if strings.HasPrefix(slashPath, "templates/") || strings.HasPrefix(slashPath, "partials/") || slashPath == "templates" || slashPath == "partials" {
            return true
        }
        return false
    }

    // File rules
    if strings.HasPrefix(slashPath, "templates/static/") {
		ext := strings.ToLower(filepath.Ext(slashPath))
		switch ext {
		case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
			return true
		}
	} else if strings.HasPrefix(slashPath, "templates/") {
		if strings.HasSuffix(slashPath, ".html") {
			return true
		}
	} else if strings.HasPrefix(slashPath, "partials/") {
		if strings.HasSuffix(slashPath, ".html") {
			return true
		}
	}
    
    return false
}
