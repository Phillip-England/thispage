package cmd

import (
	"fmt"
	"github.com/phillip-england/thispage/pkg/project"
	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init <project-name>",
	Short: "Initialize a new thispage project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		err := project.New(projectName, forceInit)
		if err != nil {
			fmt.Printf("Error initializing project: %v\n", err)
			return
		}
		fmt.Printf("Successfully initialized project '%s'\n", projectName)
	},
}

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Overwrite existing directory")
	rootCmd.AddCommand(initCmd)
}