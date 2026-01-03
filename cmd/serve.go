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
		portFromArgs := ""

		if len(args) == 1 {
			if isValidPort(args[0]) {
				portFromArgs = args[0]
			} else {
				projectPath = args[0]
			}
		}

		if len(args) == 2 {
			projectPath = args[0]
			if !isValidPort(args[1]) {
				log.Fatalf("Invalid port: %s", args[1])
			}
			portFromArgs = args[1]
		}

		if port != "" && !isValidPort(port) {
			log.Fatalf("Invalid port: %s", port)
		}

		if port != "" && portFromArgs != "" && port != portFromArgs {
			log.Printf("Using --port %s instead of positional port %s", port, portFromArgs)
		}

		resolvedPort := port
		if resolvedPort == "" {
			resolvedPort = portFromArgs
		}
		if resolvedPort == "" {
			resolvedPort = "8080"
		}

		fmt.Println("Building project...")
		if err := compiler.Build(projectPath); err != nil {
			log.Fatalf("Error building project: %v", err)
		}
		fmt.Println("Project built successfully!")

		go watcher.Start(projectPath)

		err := server.Serve(projectPath, resolvedPort)
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
	serveCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run the server on")
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
