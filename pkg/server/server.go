package server

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/phillip-england/thispage/pkg/auth"
	"github.com/phillip-england/thispage/pkg/database"
	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/thispage/pkg/routes"
	"github.com/phillip-england/vii/vii"
)

func Serve(projectPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	absProjectPath := filepath.Join(cwd, projectPath)
	liveDirPath := filepath.Join(absProjectPath, "live")

    // Load .env
    _ = godotenv.Load(filepath.Join(absProjectPath, ".env"))

    // Init Database
    if err := database.Init(absProjectPath); err != nil {
        return fmt.Errorf("failed to init database: %w", err)
    }

    // Check for Tailwind CSS
    if _, err := exec.LookPath("tailwindcss"); err == nil {
        fmt.Println("Tailwind CSS found. Starting watch process...")
        inputPath := filepath.Join(absProjectPath, "static", "input.css")
        outputPath := filepath.Join(absProjectPath, "static", "output.css")
        
        cmd := exec.Command("tailwindcss", "-i", inputPath, "-o", outputPath, "--watch")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        
        go func() {
            if err := cmd.Run(); err != nil {
                fmt.Printf("Tailwind CSS process exited with error: %v\n", err)
            }
        }()
    } else {
        fmt.Println("WARNING: Tailwind CSS not found.")
        fmt.Println("To enable automatic CSS compilation, please install the standalone executable:")
        fmt.Println("https://tailwindcss.com/blog/standalone-cli")
    }
	
	app := vii.NewApp()
	app.SetContext(keys.ProjectPath, absProjectPath)
	
	if err := app.LoadTemplates("./templates", nil); err != nil {
		return err
	}

	app.Use(vii.Logger)
	
	// Serve User Project Static Files
	app.ServeDir("/static", filepath.Join(absProjectPath, "static"))
	
	// Serve Admin Interface Static Files (from tool root)
	app.ServeDir("/admin/assets", filepath.Join(cwd, "static"))

    // Custom handler for live directory to support clean URLs (extensionless .html)
    app.Handle("GET /", func(w http.ResponseWriter, r *http.Request) {
        isAuthenticated := auth.IsAuthenticated(r)
        isAdminParam := r.URL.Query().Get("is_admin") == "true"

        // If trying to access admin mode but not authenticated, strip param
        if isAdminParam && !isAuthenticated {
             q := r.URL.Query()
             q.Del("is_admin")
             r.URL.RawQuery = q.Encode()
             http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
             return
        }

        // If authenticated but missing param, add it
        if isAuthenticated && !isAdminParam {
             q := r.URL.Query()
             q.Set("is_admin", "true")
             r.URL.RawQuery = q.Encode()
             http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
             return
        }

        urlPath := r.URL.Path
        // If path is just "/", serve index.html (handled by http.ServeFile usually, but let's be explicit or safe)
        
        fsPath := filepath.Join(liveDirPath, urlPath)

        // 1. Check if exact path exists
        info, err := os.Stat(fsPath)
        if err == nil {
            if info.IsDir() {
                // If directory, try index.html
                indexPath := filepath.Join(fsPath, "index.html")
                if _, err := os.Stat(indexPath); err == nil {
                    http.ServeFile(w, r, indexPath)
                    return
                }
                // If no index.html, 404 or list dir (let's 404 for security)
                http.NotFound(w, r)
                return
            }
            // It's a file, serve it
            http.ServeFile(w, r, fsPath)
            return
        }

        // 2. Check if path + .html exists
        htmlPath := fsPath + ".html"
        if _, err := os.Stat(htmlPath); err == nil {
            http.ServeFile(w, r, htmlPath)
            return
        }

        // 3. Not found
        http.NotFound(w, r)
    })

    authMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            if !auth.IsAuthenticated(r) {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }
            next(w, r)
        }
    }

	app.Handle("GET /login", routes.GetLogin)
	app.Handle("POST /login", routes.PostLogin)
	app.Handle("GET /admin", authMiddleware(routes.GetAdminFiles))
	app.Handle("GET /admin/files/view", authMiddleware(routes.GetAdminFileView))
	app.Handle("POST /admin/files/save", authMiddleware(routes.PostAdminFileSave))
	app.Handle("POST /admin/files/upload", authMiddleware(routes.PostAdminFileUpload))
	app.Handle("POST /admin/files/delete", authMiddleware(routes.PostAdminFileDelete))
	app.Handle("POST /admin/files/rename", authMiddleware(routes.PostAdminFileRename))
	app.Handle("POST /admin/files/create", authMiddleware(routes.PostAdminFileCreate))
	app.Handle("POST /admin/files/create-dir", authMiddleware(routes.PostAdminDirCreate))
    
    // API Routes
    app.Handle("GET /admin/api/components", authMiddleware(routes.GetAdminComponents))
    
	app.Handle("GET /admin/logout", routes.GetAdminLogout)

	return app.Serve("8080")
}


