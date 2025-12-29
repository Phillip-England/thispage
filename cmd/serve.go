package cmd

import (
	"fmt"
	"log"

	"github.com/phillip-england/thispage/pkg/compiler"
	"github.com/phillip-england/thispage/pkg/server"
	"github.com/phillip-england/thispage/pkg/watcher"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve <project-path>",
	Short: "Build and serve the thispage project with live reloading",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]

		fmt.Println("Building project...")
		if err := compiler.Build(projectPath); err != nil {
			log.Fatalf("Error building project: %v", err)
		}
		fmt.Println("Project built successfully!")

		go watcher.Start(projectPath)

		err := server.Serve(projectPath)
		if err != nil {
			log.Fatalf("Error serving project: %v", err)
		}

		if err := compiler.Build(projectPath); err != nil {
			log.Fatalf("Error building project: %v", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
