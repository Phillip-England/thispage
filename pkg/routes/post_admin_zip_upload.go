package routes

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/phillip-england/thispage/pkg/compiler"
	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/thispage/pkg/tailwind"
	"github.com/phillip-england/vii/vii"
)

func PostAdminZipUpload(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	log.Printf("[ZIP DEPLOY] Starting deploy to project: %s", projectPath)

	// 50 MB limit for zip files
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		log.Printf("[ZIP DEPLOY] Error parsing form: %v", err)
		vii.WriteError(w, http.StatusBadRequest, "File too large or invalid form: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("zipfile")
	if err != nil {
		log.Printf("[ZIP DEPLOY] Error retrieving file: %v", err)
		vii.WriteError(w, http.StatusBadRequest, "Error retrieving zip file: "+err.Error())
		return
	}
	defer file.Close()

	log.Printf("[ZIP DEPLOY] Received file: %s (size: %d bytes)", handler.Filename, handler.Size)

	// Validate it's a zip file
	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".zip") {
		vii.WriteError(w, http.StatusBadRequest, "File must be a .zip file")
		return
	}

	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "thispage-zip-*")
	if err != nil {
		log.Printf("[ZIP DEPLOY] Failed to create temp dir: %v", err)
		vii.WriteError(w, http.StatusInternalServerError, "Failed to create temp directory: "+err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	log.Printf("[ZIP DEPLOY] Created temp dir: %s", tempDir)

	// Save uploaded zip to temp file
	tempZipPath := filepath.Join(tempDir, "upload.zip")
	tempZipFile, err := os.Create(tempZipPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to create temp file: "+err.Error())
		return
	}

	bytesWritten, err := io.Copy(tempZipFile, file)
	if err != nil {
		tempZipFile.Close()
		vii.WriteError(w, http.StatusInternalServerError, "Failed to save zip file: "+err.Error())
		return
	}
	tempZipFile.Close()

	log.Printf("[ZIP DEPLOY] Saved %d bytes to temp file", bytesWritten)

	// Open and extract zip
	zipReader, err := zip.OpenReader(tempZipPath)
	if err != nil {
		log.Printf("[ZIP DEPLOY] Failed to open zip: %v", err)
		vii.WriteError(w, http.StatusBadRequest, "Failed to open zip file: "+err.Error())
		return
	}
	defer zipReader.Close()

	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to create extraction directory: "+err.Error())
		return
	}

	log.Printf("[ZIP DEPLOY] Extracting %d files from zip...", len(zipReader.File))

	// Extract all files
	for _, f := range zipReader.File {
		// Clean the file name - remove leading slashes and clean path
		cleanName := filepath.Clean(f.Name)
		cleanName = strings.TrimPrefix(cleanName, "/")
		cleanName = strings.TrimPrefix(cleanName, "\\")

		// Skip empty names or current directory
		if cleanName == "" || cleanName == "." {
			continue
		}

		destPath := filepath.Join(extractDir, cleanName)

		// Security: prevent zip slip - ensure destPath is within extractDir
		absExtractDir, _ := filepath.Abs(extractDir)
		absDestPath, _ := filepath.Abs(destPath)
		if !strings.HasPrefix(absDestPath, absExtractDir) {
			log.Printf("[ZIP DEPLOY] Zip slip attempt detected: %s", f.Name)
			vii.WriteError(w, http.StatusBadRequest, "Invalid file path in zip (zip slip attempt)")
			return
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			vii.WriteError(w, http.StatusInternalServerError, "Failed to create directory: "+err.Error())
			return
		}

		// Extract file
		srcFile, err := f.Open()
		if err != nil {
			vii.WriteError(w, http.StatusInternalServerError, "Failed to open file in zip: "+err.Error())
			return
		}

		dstFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			vii.WriteError(w, http.StatusInternalServerError, "Failed to create file: "+err.Error())
			return
		}

		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			vii.WriteError(w, http.StatusInternalServerError, "Failed to extract file: "+err.Error())
			return
		}
	}

	// List what was extracted for debugging
	extractedItems, _ := os.ReadDir(extractDir)
	log.Printf("[ZIP DEPLOY] Extracted items in root: %d", len(extractedItems))
	for _, item := range extractedItems {
		log.Printf("[ZIP DEPLOY]   - %s (dir: %v)", item.Name(), item.IsDir())
	}

	// Find the root of the thispage project in the extracted files
	// It could be directly in extractDir, or in a subdirectory (common when zipping a folder)
	projectRoot := findThispageRoot(extractDir)
	if projectRoot == "" {
		log.Printf("[ZIP DEPLOY] Could not find valid thispage project in zip")
		vii.WriteError(w, http.StatusBadRequest, "Zip does not contain a valid thispage project (missing templates/, components/, layouts/, or static/ directories)")
		return
	}

	log.Printf("[ZIP DEPLOY] Found project root: %s", projectRoot)

	// Directories to replace (these are the content directories)
	dirsToReplace := []string{"templates", "components", "layouts", "static"}

	// Remove existing directories and copy new ones
	for _, dir := range dirsToReplace {
		srcDir := filepath.Join(projectRoot, dir)
		dstDir := filepath.Join(projectPath, dir)

		// Check if source directory exists in zip
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			log.Printf("[ZIP DEPLOY] Skipping %s (not in zip)", dir)
			continue
		}

		log.Printf("[ZIP DEPLOY] Replacing %s...", dir)

		// Remove existing directory
		if err := os.RemoveAll(dstDir); err != nil {
			log.Printf("[ZIP DEPLOY] Failed to remove %s: %v", dir, err)
			vii.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove existing %s directory: %v", dir, err))
			return
		}

		// Copy new directory
		if err := copyDir(srcDir, dstDir); err != nil {
			log.Printf("[ZIP DEPLOY] Failed to copy %s: %v", dir, err)
			vii.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to copy %s directory: %v", dir, err))
			return
		}

		log.Printf("[ZIP DEPLOY] Successfully replaced %s", dir)
	}

	// Build Tailwind CSS first (one-time build)
	log.Printf("[ZIP DEPLOY] Building Tailwind CSS...")
	if err := tailwind.BuildOnce(projectPath); err != nil {
		log.Printf("[ZIP DEPLOY] Tailwind build warning: %v", err)
		// Don't fail on Tailwind errors - the project might not use Tailwind
	}

	// Restart Tailwind watch process
	log.Printf("[ZIP DEPLOY] Restarting Tailwind watch process...")
	if err := tailwind.RestartWatch(); err != nil {
		log.Printf("[ZIP DEPLOY] Tailwind restart warning: %v", err)
		// Don't fail on Tailwind errors
	}

	// Trigger template rebuild
	log.Printf("[ZIP DEPLOY] Rebuilding templates...")
	if err := compiler.Build(projectPath); err != nil {
		log.Printf("[ZIP DEPLOY] Rebuild failed: %v", err)
		vii.WriteError(w, http.StatusInternalServerError, "Project updated but rebuild failed: "+err.Error())
		return
	}

	log.Printf("[ZIP DEPLOY] Deploy completed successfully!")
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// findThispageRoot looks for a directory containing templates/, components/, layouts/, or static/
// It searches recursively up to maxDepth levels deep
func findThispageRoot(extractDir string) string {
	log.Printf("[ZIP DEPLOY] Looking for thispage project in: %s", extractDir)
	return findThispageRootRecursive(extractDir, 0, 5) // Search up to 5 levels deep
}

func findThispageRootRecursive(dir string, depth int, maxDepth int) string {
	if depth > maxDepth {
		return ""
	}

	// Check if this directory is a thispage project
	if isThispageProject(dir) {
		log.Printf("[ZIP DEPLOY] Found project at depth %d: %s", depth, dir)
		return dir
	}

	// Check subdirectories
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories and common non-project directories
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "__MACOSX" || name == "node_modules" {
				continue
			}

			subDir := filepath.Join(dir, name)
			log.Printf("[ZIP DEPLOY] Checking subdirectory at depth %d: %s", depth+1, name)

			result := findThispageRootRecursive(subDir, depth+1, maxDepth)
			if result != "" {
				return result
			}
		}
	}

	return ""
}

// isThispageProject checks if a directory looks like a thispage project
func isThispageProject(dir string) bool {
	// Must have at least one of these directories
	requiredDirs := []string{"templates", "components", "layouts", "static"}
	foundCount := 0
	var foundDirs []string

	for _, d := range requiredDirs {
		if info, err := os.Stat(filepath.Join(dir, d)); err == nil && info.IsDir() {
			foundCount++
			foundDirs = append(foundDirs, d)
		}
	}

	log.Printf("[ZIP DEPLOY] isThispageProject(%s): found %d dirs: %v", filepath.Base(dir), foundCount, foundDirs)

	// Require at least 2 of the expected directories to be present
	return foundCount >= 2
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Printf("[ZIP DEPLOY] copyDir: failed to stat src %s: %v", src, err)
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		log.Printf("[ZIP DEPLOY] copyDir: failed to create dst %s: %v", dst, err)
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		log.Printf("[ZIP DEPLOY] copyDir: failed to read src %s: %v", src, err)
		return err
	}

	log.Printf("[ZIP DEPLOY] copyDir: copying %d items from %s to %s", len(entries), filepath.Base(src), filepath.Base(dst))

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			log.Printf("[ZIP DEPLOY] copyFile: %s", entry.Name())
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
