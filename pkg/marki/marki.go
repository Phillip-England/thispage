package marki

import (
	"fmt"
	"os"
	"os/exec"
)

// Package is the Go module path for marki
const Package = "github.com/phillip-england/marki@latest"

// IsInstalled checks if marki is available in PATH
func IsInstalled() bool {
	_, err := exec.LookPath("marki")
	return err == nil
}

// Install runs go install to install marki
func Install() error {
	fmt.Println("Installing marki...")

	cmd := exec.Command("go", "install", Package)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install marki: %w", err)
	}

	fmt.Println("marki installed successfully")
	return nil
}

// EnsureInstalled checks if marki is installed, and installs it if not
func EnsureInstalled() error {
	if IsInstalled() {
		fmt.Println("marki found in PATH")
		return nil
	}

	fmt.Println("marki not found. Installing...")
	return Install()
}

// GetPath returns the path to the marki binary
func GetPath() (string, error) {
	path, err := exec.LookPath("marki")
	if err != nil {
		return "", fmt.Errorf("marki not found in PATH: %w", err)
	}
	return path, nil
}
