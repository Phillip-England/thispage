package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/thispage/pkg/tokenizer"
	"github.com/phillip-england/vii/vii"
)

func PostAdminInsertComponent(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	targetFile := r.FormValue("target_file")
	componentName := r.FormValue("component_name")
    containerIndexStr := r.FormValue("container_token_index")

	if targetFile == "" || componentName == "" {
		vii.WriteError(w, http.StatusBadRequest, "Target file and component name are required")
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
	insertion := fmt.Sprintf("\n    {{ include ./components/%s }}\n", componentName)
    
    if containerIndexStr != "" {
        // Insert into specific container layout
        idx, err := strconv.Atoi(containerIndexStr)
        if err == nil {
            tokens := tokenizer.Tokenize(content)
            if idx >= 0 && idx < len(tokens) {
                // Find the LAYOUT token
                layoutToken := tokens[idx]
                if layoutToken.Type == tokenizer.LAYOUT {
                    // Scan forward for BLOCK "content" (or just first block?)
                    // Assume container uses "content" block
                    insertPos := -1
                    depth := 1
                    for i := idx + 1; i < len(tokens); i++ {
                        t := tokens[i]
                        if t.Type == tokenizer.LAYOUT {
                            depth++
                        } else if t.Type == tokenizer.ENDLAYOUT {
                            depth--
                            if depth == 0 {
                                break // End of layout
                            }
                        }
                        
                        if depth == 1 && t.Type == tokenizer.BLOCK {
                            // Found a block in this layout
                            // Check if it's "content"? Or just use the first block?
                            // Let's assume "content" for now as per project.go default
                            if strings.Contains(t.Content, "content") {
                                // Find ENDBLOCK for this block
                                blockDepth := 1
                                for k := i + 1; k < len(tokens); k++ {
                                    bt := tokens[k]
                                    if bt.Type == tokenizer.BLOCK {
                                        blockDepth++
                                    } else if bt.Type == tokenizer.ENDBLOCK {
                                        blockDepth--
                                        if blockDepth == 0 {
                                            insertPos = bt.Start
                                            break
                                        }
                                    }
                                }
                                if insertPos != -1 {
                                    break
                                }
                            }
                        }
                    }
                    
                    if insertPos != -1 {
                        content = content[:insertPos] + insertion + content[insertPos:]
                        goto Save
                    }
                }
            }
        }
    }
	
    // Default: Append to main block or file end
    {
        idx := strings.LastIndex(content, "{{ endblock }}")
        if idx != -1 {
            content = content[:idx] + insertion + content[idx:]
        } else {
            content = content + insertion
        }
    }

Save:
    if err := os.WriteFile(absTarget, []byte(content), 0644); err != nil {
        vii.WriteError(w, http.StatusInternalServerError, "Failed to write file: "+err.Error())
        return
    }

    w.WriteHeader(http.StatusOK)
}
