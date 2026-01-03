package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phillip-england/thispage/pkg/credentials"
	"github.com/spf13/cobra"
)

var mountCmd = &cobra.Command{
	Use:   "mount <project-path> <username> <password>",
	Short: "Mount an existing thispage project with new credentials",
	Long: `Mount an existing thispage project by generating new credentials.

This is useful when you clone a thispage project from git, which won't
have credentials (since they are gitignored). The mount command creates
a new seed and credentials file for the project.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]
		username := args[1]
		password := args[2]

		// Verify the project path exists
		info, err := os.Stat(projectPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Error: project path '%s' does not exist\n", projectPath)
				return
			}
			fmt.Printf("Error: could not access project path: %v\n", err)
			return
		}
		if !info.IsDir() {
			fmt.Printf("Error: '%s' is not a directory\n", projectPath)
			return
		}

		// Verify it looks like a thispage project by checking for expected directories
		expectedDirs := []string{"templates", "components", "layouts", "static"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(projectPath, dir)
			if _, err := os.Stat(dirPath); os.IsNotExist(err) {
				fmt.Printf("Error: '%s' does not appear to be a thispage project (missing %s directory)\n", projectPath, dir)
				return
			}
		}

		// Ensure .thispage directory exists
		thispageDir := filepath.Join(projectPath, ".thispage")
		if err := os.MkdirAll(thispageDir, 0700); err != nil {
			fmt.Printf("Error: could not create .thispage directory: %v\n", err)
			return
		}

		// Remove existing seed and credentials if they exist (to generate fresh ones)
		seedPath := filepath.Join(thispageDir, "seed")
		credPath := filepath.Join(thispageDir, "credentials.json")
		os.Remove(seedPath)
		os.Remove(credPath)

		// Save new credentials (this will generate a new seed)
		if err := credentials.Save(projectPath, username, password); err != nil {
			fmt.Printf("Error: could not save credentials: %v\n", err)
			return
		}

		fmt.Printf("Successfully mounted project '%s' with new credentials\n", projectPath)
	},
}

func init() {
	rootCmd.AddCommand(mountCmd)
}
