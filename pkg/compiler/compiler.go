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

            // Inject data-source-path into body
            dataSourcePath := filepath.Join("templates", relativePath)
            if strings.Contains(compiledContent, "<body") {
                compiledContent = strings.Replace(compiledContent, "<body", fmt.Sprintf("<body data-source-path=\"%s\"", dataSourcePath), 1)
            }

            // Inject Admin Mode Script
            adminScript := `
<script>
  (function() {
    const params = new URLSearchParams(window.location.search);
    if (params.get('is_admin') === 'true') {
      // 1. Inject Styles for Blocks
      const style = document.createElement('style');
      style.textContent = 
        ".thispage-block:hover {" +
            "outline: 2px solid #3b82f6;" +
            "cursor: pointer;" +
            "position: relative;" +
        "}" +
        ".thispage-block:hover::after {" +
            "content: 'Edit';" +
            "position: absolute;" +
            "top: -20px;" +
            "right: 0;" +
            "background: #3b82f6;" +
            "color: white;" +
            "font-size: 10px;" +
            "padding: 2px 6px;" +
            "border-radius: 2px;" +
            "font-family: sans-serif;" +
            "pointer-events: none;" +
        "}";
      document.head.appendChild(style);

      // 2. Toggle Admin UI Elements
      document.querySelectorAll('.thispage-add-btn').forEach(el => el.classList.remove('hidden'));
      document.querySelectorAll('.empty-state').forEach(el => el.style.display = 'none');

      // 3. Inject Back to Admin Button
      const sourcePath = document.body.getAttribute('data-source-path');
      if (sourcePath) {
          const backBtn = document.createElement('a');
          backBtn.href = '/admin/files/view?path=' + encodeURIComponent(sourcePath);
          backBtn.textContent = 'Back to Edit';
          backBtn.style.cssText = 'position:fixed;bottom:2rem;left:2rem;z-index:9999;background:#171717;color:white;padding:0.5rem 1rem;border-radius:0.375rem;font-family:sans-serif;font-size:0.875rem;text-decoration:none;border:1px solid #404040;box-shadow:0 4px 6px -1px rgba(0,0,0,0.1);';
          backBtn.onmouseover = () => backBtn.style.background = '#262626';
          backBtn.onmouseout = () => backBtn.style.background = '#171717';
          document.body.appendChild(backBtn);
      }

      // 4. Intercept Links to Persist Admin State
      document.addEventListener('click', function(e) {
        // Add Button Handler
        if (e.target.closest('.thispage-add-btn')) {
             e.preventDefault();
             e.stopPropagation();
             
             // Fetch partials
             fetch('/admin/api/partials')
                .then(res => res.json())
                .then(data => {
                    // Create and show modal
                    let modal = document.getElementById('partial-modal');
                    if (!modal) {
                        modal = document.createElement('div');
                        modal.id = 'partial-modal';
                        modal.style.cssText = 'position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.8);z-index:9999;display:flex;align-items:center;justify-content:center;';
                        
                        const content = document.createElement('div');
                        content.style.cssText = 'background:#171717;border:1px solid #404040;padding:24px;border-radius:8px;width:300px;color:white;';
                        
                        const title = document.createElement('h3');
                        title.textContent = 'Insert Block';
                        title.style.cssText = 'font-weight:bold;margin-bottom:16px;text-transform:uppercase;letter-spacing:1px;font-size:12px;color:#a3a3a3;';
                        content.appendChild(title);
                        
                        const list = document.createElement('ul');
                        list.id = 'partial-list';
                        content.appendChild(list);

                        const close = document.createElement('button');
                        close.textContent = 'Cancel';
                        close.style.cssText = 'margin-top:16px;width:100%;padding:8px;background:transparent;border:1px solid #404040;color:#a3a3a3;cursor:pointer;font-size:12px;text-transform:uppercase;letter-spacing:1px;';
                        close.onclick = () => modal.style.display = 'none';
                        content.appendChild(close);
                        
                        modal.appendChild(content);
                        document.body.appendChild(modal);
                    }
                    
                    const list = document.getElementById('partial-list');
                    list.innerHTML = ''; // clear
                    
                    data.files.forEach(file => {
                        const item = document.createElement('li');
                        item.textContent = file;
                        item.style.cssText = 'padding:12px;background:#262626;margin-bottom:8px;cursor:pointer;border-radius:4px;font-size:14px;';
                        item.onmouseover = () => item.style.background = '#404040';
                        item.onmouseout = () => item.style.background = '#262626';
                        
                        item.onclick = () => {
                             const sourcePath = document.body.getAttribute('data-source-path');
                             fetch('/admin/api/insert-partial', {
                                method: 'POST',
                                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                                body: 'target_file=' + encodeURIComponent(sourcePath) + '&partial_name=' + encodeURIComponent('partials/' + file)
                             }).then(res => {
                                 if (res.ok) {
                                     window.location.reload();
                                 } else {
                                     alert('Error inserting partial');
                                 }
                             });
                        };
                        list.appendChild(item);
                    });
                    
                    modal.style.display = 'flex';
                });
             return;
        }

        // Block Click Handler
        if (e.target.classList.contains('thispage-block')) {
            e.preventDefault();
            e.stopPropagation();
            console.log("Block clicked:", e.target);
            alert("Edit block functionality coming soon!");
            return;
        }

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
