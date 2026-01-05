package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate <source-project> <target-project>",
	Short: "Migrate credentials and database from one project to another",
	Long: `Migrate copies the .thispage directory (credentials) and data.db (database)
from a source project to a target project. This is useful when updating to a new
version of your application while preserving your existing data.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		target := args[1]

		if err := migrateProject(source, target); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migration completed successfully!")
	},
}

func migrateProject(source, target string) error {
	// Validate source exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source project '%s' does not exist", source)
	}

	// Validate target exists
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("target project '%s' does not exist", target)
	}

	// Copy .thispage directory (credentials)
	sourceThispage := filepath.Join(source, ".thispage")
	targetThispage := filepath.Join(target, ".thispage")

	if _, err := os.Stat(sourceThispage); err == nil {
		// Remove existing .thispage in target if present
		os.RemoveAll(targetThispage)

		if err := copyDir(sourceThispage, targetThispage); err != nil {
			return fmt.Errorf("failed to copy credentials: %w", err)
		}
		fmt.Println("Copied credentials (.thispage/)")
	} else {
		fmt.Println("Warning: No .thispage directory found in source, skipping credentials")
	}

	// Copy data.db
	sourceDB := filepath.Join(source, "data.db")
	targetDB := filepath.Join(target, "data.db")

	if _, err := os.Stat(sourceDB); err == nil {
		// Remove existing data.db in target if present
		os.Remove(targetDB)

		if err := copyFile(sourceDB, targetDB); err != nil {
			return fmt.Errorf("failed to copy database: %w", err)
		}
		fmt.Println("Copied database (data.db)")
	} else {
		fmt.Println("Warning: No data.db found in source, skipping database")
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
