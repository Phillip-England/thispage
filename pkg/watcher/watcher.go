package watcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/phillip-england/thispage/pkg/compiler"
)

func Start(projectPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("failed to create watcher: %v\n", err)
		return
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("Changes detected, rebuilding site...")
					if err := compiler.Build(projectPath); err != nil {
						fmt.Printf("Error rebuilding site: %v\n", err)
					} else {
						fmt.Println("Site rebuilt successfully!")
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	for _, dir := range []string{"templates", "partials"} {
		path := filepath.Join(projectPath, dir)
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					fmt.Printf("failed to add path to watcher: %v\n", err)
				}
			}
			return nil
		}); err != nil {
			fmt.Printf("failed to walk directory %s: %v\n", path, err)
		}
	}

	fmt.Printf("Watching for changes in %s\n", projectPath)
}
