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
	Link     string // URL to navigate to when clicked (edit), empty if not clickable
	PageLink string // URL to view the rendered page (templates only)
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
	allowedDirs := []string{"components", "templates", "static", "layouts", "content"}
    
    var directories []string

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
        
        directories = append(directories, dirName)

		if err := buildTree(absDir, dirNode, dirName, &directories); err != nil {
			vii.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		root.Children = append(root.Children, dirNode)
	}
    
    // Sort directories for better UX
    sort.Strings(directories)

	err := vii.Render(w, r, "admin_files.html", map[string]interface{}{
		"Files":       root,
		"ProjectPath": projectPath,
        "Directories": directories,
	})
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, err.Error())
	}
}

func buildTree(absPath string, node *FileNode, rootDir string, dirs *[]string) error {
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
		// Skip hidden files/dirs
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		relPath := filepath.ToSlash(filepath.Join(node.Path, entry.Name()))

		// Filter logic
		if !entry.IsDir() {
			if !isAllowedFile(relPath) {
				continue
			}
		} else {
             *dirs = append(*dirs, relPath)
        }

		editLink, pageLink := computeLinks(relPath, entry.IsDir())
		child := &FileNode{
			Name:     entry.Name(),
			IsDir:    entry.IsDir(),
			Path:     relPath,
			Link:     editLink,
			PageLink: pageLink,
		}

		if entry.IsDir() {
			if err := buildTree(filepath.Join(absPath, entry.Name()), child, rootDir, dirs); err != nil {
				return err
			}
		}

		node.Children = append(node.Children, child)
	}
	return nil
}

func isAllowedFile(relPath string) bool {
	// Normalize separators
	relPath = filepath.ToSlash(relPath)

	// Check for static directory (top level)
	if strings.HasPrefix(relPath, "static/") {
		ext := strings.ToLower(filepath.Ext(relPath))
		switch ext {
		case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
			return true
		}
		return false
	}

	// Check for templates directory
	if strings.HasPrefix(relPath, "templates/") {
		return strings.HasSuffix(relPath, ".html")
	}

	// Check for components directory
	if strings.HasPrefix(relPath, "components/") {
		return strings.HasSuffix(relPath, ".html")
	}

    // Check for layouts directory
    if strings.HasPrefix(relPath, "layouts/") {
        return strings.HasSuffix(relPath, ".html")
    }

	// Check for content directory
	if strings.HasPrefix(relPath, "content/") {
		return strings.HasSuffix(relPath, ".md")
	}

	return false
}

func computeLinks(relPath string, isDir bool) (editLink, pageLink string) {
	if isDir {
		return "", ""
	}

	// Edit link for all files
	editLink = "/admin/files/view?path=" + relPath

	// Page link only for templates (viewable pages)
	if strings.HasPrefix(relPath, "templates/") && strings.HasSuffix(relPath, ".html") {
		// Convert templates/foo/bar.html -> /foo/bar?is_admin=true
		pagePath := strings.TrimPrefix(relPath, "templates")
		pagePath = strings.TrimSuffix(pagePath, ".html")
		if pagePath == "/index" {
			pagePath = "/"
		}
		pageLink = pagePath + "?is_admin=true"
	}

	return editLink, pageLink
}