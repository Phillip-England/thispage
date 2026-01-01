package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
    "strings"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

type FileNode struct {
	Name     string
	IsDir    bool
	Path     string
	Children []*FileNode
}

func GetAdminFiles(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
    if !ok {
        vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
        return
    }

	root := &FileNode{
		Name:  filepath.Base(projectPath),
		IsDir: true,
		Path:  "/",
	}

	// Only allow specific top-level directories
	allowedDirs := []string{"partials", "templates", "static"}

	for _, dirName := range allowedDirs {
		absDir := filepath.Join(projectPath, dirName)
		
		// Check if directory exists
		info, err := os.Stat(absDir)
		if err != nil || !info.IsDir() {
			continue
		}

		dirNode := &FileNode{
			Name:  dirName,
			IsDir: true,
			Path:  dirName, // Relative path for links
		}

		if err := buildTree(absDir, dirNode); err != nil {
			vii.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		root.Children = append(root.Children, dirNode)
	}

	err := vii.Render(w, r, "admin_files.html", map[string]interface{}{
		"Files": root,
        "ProjectPath": projectPath,
	})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}

func buildTree(absPath string, node *FileNode) error {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return err
	}

	// Sort entries: directories first, then files
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		}
		if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
        // Skip hidden files/dirs (optional, but usually good)
        if strings.HasPrefix(entry.Name(), ".") {
            continue
        }

		child := &FileNode{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			// Constructing a relative path or using the name; for now, let's store the name.
            // If we need a clickable path, we might need to build it up.
            // Let's assume for display we just need structure.
			Path:  filepath.ToSlash(filepath.Join(node.Path, entry.Name())), 
		}

		if entry.IsDir() {
			if err := buildTree(filepath.Join(absPath, entry.Name()), child); err != nil {
				return err
			}
		}

		node.Children = append(node.Children, child)
	}
	return nil
}
