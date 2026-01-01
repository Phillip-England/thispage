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

            // Inject Admin Mode Script
            adminScript := `
<script>
  (function() {
    const params = new URLSearchParams(window.location.search);
    if (params.get('is_admin') === 'true') {
      // 1. Show Badge
      const badge = document.createElement('div');
      badge.textContent = 'Admin Mode';
      badge.style.position = 'fixed';
      badge.style.bottom = '10px';
      badge.style.right = '10px';
      badge.style.backgroundColor = '#ef4444';
      badge.style.color = 'white';
      badge.style.padding = '4px 8px';
      badge.style.borderRadius = '4px';
      badge.style.fontSize = '12px';
      badge.style.fontFamily = 'sans-serif';
      badge.style.zIndex = '9999';
      badge.style.pointerEvents = 'none';
      badge.style.opacity = '0.8';
      document.body.appendChild(badge);

      // 2. Intercept Links to Persist Admin State
      document.addEventListener('click', function(e) {
        const link = e.target.closest('a');
        if (link && link.href) {
            const url = new URL(link.href, window.location.origin);
            // Only modify internal links
            if (url.origin === window.location.origin) {
                // specific check to avoid duplicating or messing up non-nav links
                if (!url.searchParams.has('is_admin')) {
                    url.searchParams.set('is_admin', 'true');
                    link.href = url.toString();
                }
            }
        }
      });
    }
  })();
</script>`
            if strings.Contains(compiledContent, "</body>") {
                compiledContent = strings.Replace(compiledContent, "</body>", adminScript+"</body>", 1)
            } else {
                compiledContent += adminScript
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
