package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/phillip-england/thispage/pkg/compiler"
	"github.com/phillip-england/thispage/pkg/server"
	"github.com/phillip-england/thispage/pkg/watcher"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve [project-path] [port]",
	Short: "Build and serve the thispage project with live reloading",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := "."
		port = "8080"

		if len(args) == 1 {
			if isValidPort(args[0]) {
				port = args[0]
			} else {
				projectPath = args[0]
			}
		}

		if len(args) == 2 {
			projectPath = args[0]
			if !isValidPort(args[1]) {
				log.Fatalf("Invalid port: %s", args[1])
			}
			port = args[1]
		}

		fmt.Println("Building project...")
		if err := compiler.Build(projectPath); err != nil {
			log.Fatalf("Error building project: %v", err)
		}
		fmt.Println("Project built successfully!")

		go watcher.Start(projectPath)

		err := server.Serve(projectPath, port)
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

func isValidPort(value string) bool {
	if value == "" {
		return false
	}
	portNumber, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return portNumber > 0 && portNumber <= 65535
}
