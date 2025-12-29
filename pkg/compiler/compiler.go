package compiler

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"github.com/phillip-england/thispage/pkg/tokenizer"
)

func Compile(tokens []tokenizer.Token, projectPath string) (string, error) {
	var builder strings.Builder
	for _, token := range tokens {
		switch token.Type {
		case tokenizer.RAWHTML:
			builder.WriteString(token.Content)
		case tokenizer.INCLUDE:
			includePath := filepath.Join(projectPath, token.Content)

			cleanPath, err := filepath.Abs(includePath)
			if err != nil {
				return "", fmt.Errorf("error cleaning path: %w", err)
			}

			cleanProjectPath, err := filepath.Abs(projectPath)
			if err != nil {
				return "", fmt.Errorf("error cleaning project path: %w", err)
			}

			if !strings.HasPrefix(cleanPath, cleanProjectPath) {
				return "", fmt.Errorf("include path is outside the project directory: %s", token.Content)
			}
			content, err := os.ReadFile(cleanPath)
			if err != nil {
				return "", fmt.Errorf("failed to read include file '%s': %w", includePath, err)
			}
			builder.Write(content)
		}
	}
	return builder.String(), nil
}

func Build(projectPath string) error {
	templatesPath := filepath.Join(projectPath, "templates")
	livePath := filepath.Join(projectPath, "live")
	compiledFiles := make(map[string]string)
	err := filepath.Walk(templatesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "admin" && path == filepath.Join(templatesPath, "admin") {
				return fmt.Errorf("the 'admin' directory is reserved and cannot be created in 'templates'")
			}
			return nil
		}
		if filepath.Ext(path) == ".html" {
			relativePath, err := filepath.Rel(templatesPath, path)
			if err != nil {
				return err
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			tokens := tokenizer.Tokenize(string(content))
			compiledContent, err := Compile(tokens, projectPath)
			if err != nil {
				return fmt.Errorf("error compiling %s: %w", path, err)
			}
			destPath := filepath.Join(livePath, relativePath)
			compiledFiles[destPath] = compiledContent
		}
		return nil
	})
	if err != nil {
		return err // Return the compilation error without writing any files
	}
	if err := os.RemoveAll(livePath); err != nil {
		return fmt.Errorf("failed to remove live directory: %w", err)
	}
	if err := os.MkdirAll(livePath, 0755); err != nil {
		return fmt.Errorf("failed to create live directory: %w", err)
	}
	for destPath, content := range compiledFiles {
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}
		if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}
	}
	return nil
}
