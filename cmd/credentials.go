package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phillip-england/thispage/pkg/credentials"
	"github.com/spf13/cobra"
)

var credentialsCmd = &cobra.Command{
	Use:   "credentials <project-path>",
	Short: "Show the stored admin credentials for a thispage project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, err := absPath(args[0])
		if err != nil {
			fmt.Printf("Error resolving project path: %v\n", err)
			return
		}

		username, password, err := credentials.Load(projectPath)
		if err != nil {
			fmt.Printf("Error loading credentials: %v\n", err)
			return
		}

		sessionKey, err := credentials.ProjectSeed(projectPath)
		if err != nil {
			fmt.Printf("Error loading session key: %v\n", err)
			return
		}

		fmt.Printf("Username:    %s\nPassword:    %s\nSession Key: %s\n", username, password, sessionKey)
	},
}

var credentialsSetCmd = &cobra.Command{
	Use:   "set <project-path> <username> <password>",
	Short: "Update the stored admin credentials for a thispage project",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath, err := absPath(args[0])
		if err != nil {
			fmt.Printf("Error resolving project path: %v\n", err)
			return
		}

		username := args[1]
		password := args[2]

		if err := credentials.Save(projectPath, username, password); err != nil {
			fmt.Printf("Error saving credentials: %v\n", err)
			return
		}

		fmt.Printf("Credentials updated for '%s'\n", projectPath)
	},
}

func init() {
	credentialsCmd.AddCommand(credentialsSetCmd)
	rootCmd.AddCommand(credentialsCmd)
}

func absPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}
