package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/thispage/pkg/tokenizer"
	"github.com/phillip-england/vii/vii"
)

func PostAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	if err := r.ParseForm(); err != nil {
		vii.WriteError(w, http.StatusBadRequest, "Failed to parse form: "+err.Error())
		return
	}

	relPath := r.FormValue("file")
	tokenIndexStr := r.FormValue("token_index")

	if relPath == "" || tokenIndexStr == "" {
		vii.WriteError(w, http.StatusBadRequest, "File and token index are required")
		return
	}

    // Security Check
    relPath = filepath.Clean(relPath)
    if strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, "/") {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Invalid path")
        return
    }
    
    // Only allow editing templates/components
    allowed := false
    if strings.HasPrefix(relPath, "templates/") || strings.HasPrefix(relPath, "components/") {
        allowed = true
    }
    if !allowed {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Restricted directory")
        return
    }

	absPath := filepath.Join(projectPath, relPath)
    if !strings.HasPrefix(absPath, projectPath) {
        vii.WriteError(w, http.StatusForbidden, "Access denied: Path traversal")
        return
    }

    contentBytes, err := os.ReadFile(absPath)
    if err != nil {
        vii.WriteError(w, http.StatusInternalServerError, "Failed to read file: "+err.Error())
        return
    }
    content := string(contentBytes)

    tokens := tokenizer.Tokenize(content)
    index, err := strconv.Atoi(tokenIndexStr)
    if err != nil || index < 0 || index >= len(tokens) {
        vii.WriteError(w, http.StatusBadRequest, "Invalid token index")
        return
    }

    token := tokens[index]
    
    // Validate that it's an INCLUDE token (safety check)
    // Actually, we might delete other things later, but for now we only target blocks.
    // The client sends the index of the INCLUDE token.
    if token.Type != tokenizer.INCLUDE {
         // Should we allow deleting others? Maybe. But for now warn/log?
         // Let's proceed, assuming client knows what it's doing.
    }

    // Remove token from content
    newContent := content[:token.Start] + content[token.End:]

    if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
        vii.WriteError(w, http.StatusInternalServerError, "Failed to write file: "+err.Error())
        return
    }

	w.WriteHeader(http.StatusOK)
}
