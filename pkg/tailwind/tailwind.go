package tailwind

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Version is the Tailwind CSS version to install
// Update this constant to upgrade Tailwind for all thispage projects
const Version = "4.1.18"

// BaseURL is the GitHub releases download URL pattern
const BaseURL = "https://github.com/tailwindlabs/tailwindcss/releases/download/v%s/%s"

// GetBinaryName returns the correct binary name for the current OS and architecture
func GetBinaryName() (string, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "darwin":
		switch arch {
		case "arm64":
			return "tailwindcss-macos-arm64", nil
		case "amd64":
			return "tailwindcss-macos-x64", nil
		default:
			return "", fmt.Errorf("unsupported macOS architecture: %s", arch)
		}
	case "linux":
		switch arch {
		case "arm64":
			return "tailwindcss-linux-arm64", nil
		case "amd64":
			return "tailwindcss-linux-x64", nil
		default:
			return "", fmt.Errorf("unsupported Linux architecture: %s", arch)
		}
	case "windows":
		switch arch {
		case "amd64":
			return "tailwindcss-windows-x64.exe", nil
		default:
			return "", fmt.Errorf("unsupported Windows architecture: %s", arch)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", os)
	}
}

// GetInstallDir returns the directory where tailwindcss will be installed
func GetInstallDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".thispage", "bin"), nil
}

// GetBinaryPath returns the full path to the tailwindcss binary
func GetBinaryPath() (string, error) {
	installDir, err := GetInstallDir()
	if err != nil {
		return "", err
	}

	binaryName := "tailwindcss"
	if runtime.GOOS == "windows" {
		binaryName = "tailwindcss.exe"
	}

	return filepath.Join(installDir, binaryName), nil
}

// GetVersionFilePath returns the path to the version file
func GetVersionFilePath() (string, error) {
	installDir, err := GetInstallDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(installDir, ".tailwindcss-version"), nil
}

// GetInstalledVersion reads the installed version from the version file
func GetInstalledVersion() (string, error) {
	versionPath, err := GetVersionFilePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(versionPath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// WriteVersionFile writes the current version to the version file
func WriteVersionFile() error {
	versionPath, err := GetVersionFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(versionPath, []byte(Version), 0644)
}

// IsInstalled checks if tailwindcss is already installed
func IsInstalled() bool {
	binaryPath, err := GetBinaryPath()
	if err != nil {
		return false
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

// NeedsUpdate checks if the installed version differs from the expected version
func NeedsUpdate() bool {
	if !IsInstalled() {
		return true
	}

	installedVersion, err := GetInstalledVersion()
	if err != nil {
		// No version file means we need to update
		return true
	}

	return installedVersion != Version
}

// Download downloads the tailwindcss binary for the current platform
func Download() error {
	binaryName, err := GetBinaryName()
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf(BaseURL, Version, binaryName)
	fmt.Printf("Downloading Tailwind CSS v%s...\n", Version)

	// Create install directory
	installDir, err := GetInstallDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	// Download the binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Get the target path
	binaryPath, err := GetBinaryPath()
	if err != nil {
		return err
	}

	// Create temporary file first
	tmpPath := binaryPath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy the downloaded content
	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmpFile.Close()

	fmt.Printf("Downloaded %d bytes\n", written)

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	// Move temp file to final location
	if err := os.Rename(tmpPath, binaryPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move binary to final location: %w", err)
	}

	// Write version file
	if err := WriteVersionFile(); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	fmt.Printf("Tailwind CSS v%s installed to %s\n", Version, binaryPath)
	return nil
}

// EnsureInstalled checks if the correct version of tailwindcss is installed,
// and installs/updates it if needed. Always uses the thispage-managed binary.
func EnsureInstalled() (string, error) {
	binaryPath, err := GetBinaryPath()
	if err != nil {
		return "", err
	}

	// Check if we need to install or update
	if NeedsUpdate() {
		installedVersion, _ := GetInstalledVersion()
		if installedVersion != "" {
			fmt.Printf("Tailwind CSS version mismatch (installed: %s, required: %s). Updating...\n", installedVersion, Version)
		} else {
			fmt.Println("Tailwind CSS not found. Installing...")
		}

		if err := Download(); err != nil {
			return "", err
		}
	} else {
		fmt.Printf("Tailwind CSS v%s ready\n", Version)
	}

	return binaryPath, nil
}
