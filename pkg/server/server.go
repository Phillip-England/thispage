package server

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/phillip-england/thispage/pkg/routes"
	"github.com/phillip-england/vii/vii"
)

func Serve(projectPath string) error {
	_ = projectPath // unused for now

	errCh := make(chan error, 2)

	// Admin server
	go func() {
		errCh <- ServeAdmin()
	}()

	// Static server (no-op for now, but runs side-by-side)
	go func() {
		errCh <- ServeStatic(projectPath)
	}()

	// Return when the first one finishes (or errors).
	if err := <-errCh; err != nil {
		return err
	}
	// If the first completion was nil (static no-op), wait for the admin server.
	return <-errCh
}

func ServeStatic(projectPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	liveDir := path.Join(cwd, projectPath, "live", "**", "*.html")
	fmt.Println("Attmpting to serve files using bun from: " + liveDir)
	serverCmd := exec.Command("bun", liveDir)
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	serverCmd.Stdin = os.Stdin // optional, but useful

	err = serverCmd.Run()

	if err != nil {
		return err
	}
	return nil
}

func ServeAdmin() error {
	app := vii.New()
	app.Use(vii.LoggerService{})

	if err := app.Mount(http.MethodGet, "/", routes.BunProxy{}); err != nil {
		return err
	}

	fmt.Println("Admin server on :8080")
	return http.ListenAndServe(":8080", app)
}
