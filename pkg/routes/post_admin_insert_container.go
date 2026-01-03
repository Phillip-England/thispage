package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func PostAdminInsertContainer(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	targetFile := r.FormValue("target_file")
	containerName := r.FormValue("container_name") // e.g. "flex_row.html"

	if targetFile == "" || containerName == "" {
		vii.WriteError(w, http.StatusBadRequest, "Target file and container name are required")
		return
	}

	absTarget := filepath.Join(projectPath, targetFile)
    if !strings.HasPrefix(absTarget, projectPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied")
        return
    }
	
	contentBytes, err := os.ReadFile(absTarget)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to read target file: "+err.Error())
		return
	}
    content := string(contentBytes)

    // Layout structure for container
	insertion := fmt.Sprintf("\n{{ layout (\"./containers/%s\") }}\n{{ block \"content\" }}\n{{ endblock }}\n{{ endlayout }}\n", containerName)
	
    idx := strings.LastIndex(content, "{{ endblock }}")
    if idx != -1 {
        content = content[:idx] + insertion + content[idx:]
    } else {
        content = content + insertion
    }

    if err := os.WriteFile(absTarget, []byte(content), 0644); err != nil {
        vii.WriteError(w, http.StatusInternalServerError, "Failed to write file: "+err.Error())
        return
    }

    w.WriteHeader(http.StatusOK)
}
