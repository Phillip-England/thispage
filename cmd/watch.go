package cmd

import (
	"github.com/phillip-england/thispage/pkg/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <project-path>",
	Short: "Build the thispage project and watch for changes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]
		watcher.Start(projectPath)
		select {}
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
