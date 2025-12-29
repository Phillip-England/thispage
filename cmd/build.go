package cmd

import (
	"fmt"
	"github.com/phillip-england/thispage/pkg/compiler"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build <project-path>",
	Short: "Build the thispage project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]
		fmt.Printf("Building project at %s...\n", projectPath)
		if err := compiler.Build(projectPath); err != nil {
			fmt.Printf("Error building project: %v\n", err)
			return
		}
		fmt.Println("Project built successfully!")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
