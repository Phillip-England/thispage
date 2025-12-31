package server

import (
	"os"
	"path"

	"github.com/phillip-england/thispage/pkg/routes"
	"github.com/phillip-england/vii/vii"
)

func Serve(projectPath string) error {
	_ = projectPath
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	liveDirPath := path.Join(cwd, projectPath, "live")
	app := vii.NewApp()
	
	if err := app.LoadTemplates("./templates", nil); err != nil {
		return err
	}

	app.Use(vii.Logger)
	
	app.ServeDir("/", liveDirPath)
	app.ServeDir("/static", path.Join(cwd, "static"))

	app.Handle("GET /admin/login", routes.GetAdminLogin)
	app.Handle("POST /admin/login", routes.PostAdminLogin)

	return app.Serve("8080")
}


