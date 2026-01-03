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
	"time"

	"github.com/phillip-england/thispage/pkg/keys"
	"github.com/phillip-england/vii/vii"
)

func GetAdminExport(w http.ResponseWriter, r *http.Request) {
	projectPath, ok := vii.GetContext(keys.ProjectPath, r).(string)
	if !ok {
		vii.WriteError(w, http.StatusInternalServerError, "Project path not found in context")
		return
	}

	log.Printf("[EXPORT] Starting export of project: %s", projectPath)

	// Directories to export (excludes .thispage, data.db, live)
	dirsToExport := []string{"templates", "components", "layouts", "static"}

	// Create a temporary file for the zip
	tempFile, err := os.CreateTemp("", "thispage-export-*.zip")
	if err != nil {
		log.Printf("[EXPORT] Failed to create temp file: %v", err)
		vii.WriteError(w, http.StatusInternalServerError, "Failed to create export file")
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Create zip writer
	zipWriter := zip.NewWriter(tempFile)

	// Add each directory to the zip
	for _, dir := range dirsToExport {
		srcDir := filepath.Join(projectPath, dir)

		// Check if directory exists
		if _, err := os.Stat(srcDir); os.IsNotExist(err) {
			log.Printf("[EXPORT] Skipping %s (does not exist)", dir)
			continue
		}

		log.Printf("[EXPORT] Adding directory: %s", dir)

		// Walk the directory and add files
		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get relative path from project root
			relPath, err := filepath.Rel(projectPath, path)
			if err != nil {
				return err
			}

			// Skip hidden files/directories (except the directory itself)
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") && path != srcDir {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Create zip header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Use forward slashes for zip compatibility
			header.Name = filepath.ToSlash(relPath)

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return err
			}

			// If it's a file, copy its contents
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(writer, file)
				if err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			log.Printf("[EXPORT] Error adding %s: %v", dir, err)
			zipWriter.Close()
			tempFile.Close()
			vii.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export %s: %v", dir, err))
			return
		}
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		log.Printf("[EXPORT] Failed to close zip: %v", err)
		tempFile.Close()
		vii.WriteError(w, http.StatusInternalServerError, "Failed to finalize export")
		return
	}
	tempFile.Close()

	// Get file size for Content-Length
	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to get export file info")
		return
	}

	// Generate filename with timestamp
	projectName := filepath.Base(projectPath)
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := fmt.Sprintf("%s_export_%s.zip", projectName, timestamp)

	log.Printf("[EXPORT] Export complete: %s (%d bytes)", filename, fileInfo.Size())

	// Send the file as a download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Open and stream the file
	exportFile, err := os.Open(tempPath)
	if err != nil {
		vii.WriteError(w, http.StatusInternalServerError, "Failed to read export file")
		return
	}
	defer exportFile.Close()

	io.Copy(w, exportFile)
}
