package server

import (
	"fmt"
	"net/http"
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
	app := vii.New()
	app.TemplateLocalDir("templates", "./templates", "*.html")
	app.Use(vii.LoggerService{})
	err = app.ServeLocalFiles("/", liveDirPath)
	if err != nil {
		return nil
	}
	err = app.ServeLocalFiles("/static", path.Join(cwd, "static"))
	if err != nil {
		return nil
	}
	if err := app.AddMany(map[string]vii.Route{
		"GET /admin/login": routes.GetAdminLogin{},
		"POST /admin/login": routes.PostAdminLogin{},
	}); err != nil {
		return err
	}
	fmt.Println("Admin server on :8080")
	return http.ListenAndServe(":8080", app)
}


