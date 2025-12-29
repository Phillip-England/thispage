package cmd

import (
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
		watcher.WatchAndServe(projectPath, port)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "port to serve the project on")
	rootCmd.AddCommand(serveCmd)
}
