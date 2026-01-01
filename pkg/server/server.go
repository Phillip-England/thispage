package server

import (
	"os"
	"path"
	"path/filepath"

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
	
	app := vii.NewApp()
	app.SetContext(keys.ProjectPath, absProjectPath)
	
	if err := app.LoadTemplates("./templates", nil); err != nil {
		return err
	}

	app.Use(vii.Logger)
	
	app.ServeDir("/", liveDirPath)
	app.ServeDir("/static", path.Join(cwd, "static"))

	app.Handle("GET /admin", routes.GetAdmin)
	app.Handle("POST /admin", routes.PostAdmin)
	app.Handle("GET /admin/dashboard", routes.GetAdminDashboard)
	app.Handle("GET /admin/files", routes.GetAdminFiles)
	app.Handle("GET /admin/files/view", routes.GetAdminFileView)
	app.Handle("POST /admin/files/save", routes.PostAdminFileSave)
	app.Handle("GET /admin/logout", routes.GetAdminLogout)

	return app.Serve("8080")
}


