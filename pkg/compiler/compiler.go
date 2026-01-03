package compiler

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phillip-england/thispage/pkg/tokenizer"
)

var argsRegex = regexp.MustCompile(`(?:(\w+)=["'](.*?)["'])|(?:\"(.*?)\"|'(.*?)'|(\S+))`)

func parseArgs(raw string) (string, map[string]string) {
	mainArg := ""
	props := make(map[string]string)
	matches := argsRegex.FindAllStringSubmatch(raw, -1)
	for _, match := range matches {
		if match[1] != "" {
			// Key-Value pair
			props[match[1]] = match[2]
		} else {
			// Positional argument (mainArg)
			val := match[3] // double quoted
			if val == "" {
				val = match[4] // single quoted
			}
			if val == "" {
				val = match[5] // unquoted
			}
			if mainArg == "" && val != "" {
				mainArg = val
			}
		}
	}
	return mainArg, props
}

func Compile(tokens []tokenizer.Token, projectPath string, currentFile string) (string, error) {
	return compileRecursive(tokens, projectPath, nil, nil, currentFile)
}

func compileRecursive(tokens []tokenizer.Token, projectPath string, blocks map[string]string, props map[string]string, currentFile string) (string, error) {
	var builder strings.Builder
	i := 0
	for i < len(tokens) {
		token := tokens[i]
		switch token.Type {
		case tokenizer.RAWHTML:
			builder.WriteString(token.Content)
			i++
		case tokenizer.INCLUDE:
			pathStr, newProps := parseArgs(token.Content)
			
			// Merge props
			mergedProps := make(map[string]string)
			for k, v := range props {
				mergedProps[k] = v
			}
			for k, v := range newProps {
				mergedProps[k] = v
			}

			path := filepath.Join(projectPath, pathStr)
			cleanPath, err := filepath.Abs(path)
			if err != nil {
				return "", err
			}
			cleanProject, err := filepath.Abs(projectPath)
			if err != nil {
				return "", err
			}
			if !strings.HasPrefix(cleanPath, cleanProject) {
				return "", fmt.Errorf("include path outside project: %s", pathStr)
			}

			content, err := os.ReadFile(cleanPath)
			if err != nil {
				return "", fmt.Errorf("failed to read include %s: %w", pathStr, err)
			}
			subTokens := tokenizer.Tokenize(string(content))
            
            // Output Start Marker
            builder.WriteString(fmt.Sprintf("<!-- __TP_INC__ file=\"%s\" token_index=\"%d\" -->", currentFile, i))
            
			subOutput, err := compileRecursive(subTokens, projectPath, blocks, mergedProps, pathStr)
			if err != nil {
				return "", err
			}
			builder.WriteString(subOutput)
            
            // Output End Marker
            builder.WriteString("<!-- __TP_END_INC__ -->")
            
			i++
		case tokenizer.LAYOUT:
			pathStr, newProps := parseArgs(token.Content)
			layoutPath := filepath.Join(projectPath, pathStr)
			
			// Merge props for layout
			mergedProps := make(map[string]string)
			for k, v := range props {
				mergedProps[k] = v
			}
			for k, v := range newProps {
				mergedProps[k] = v
			}

			layoutContentTokens, nextIndex := extractTokensUntil(tokens, i+1, tokenizer.LAYOUT, tokenizer.ENDLAYOUT)
			i = nextIndex 

			newBlocks := make(map[string]string)
			j := 0
			for j < len(layoutContentTokens) {
				t := layoutContentTokens[j]
				if t.Type == tokenizer.BLOCK {
					blockName, _ := parseArgs(t.Content)
					blockTokens, nextJ := extractTokensUntil(layoutContentTokens, j+1, tokenizer.BLOCK, tokenizer.ENDBLOCK)
					j = nextJ
					compiledBlock, err := compileRecursive(blockTokens, projectPath, nil, mergedProps, currentFile) 
					if err != nil {
						return "", err
					}
					newBlocks[blockName] = compiledBlock
				} else {
					j++
				}
			}

			cleanPath, err := filepath.Abs(layoutPath)
			if err != nil {
				return "", err
			}
			cleanProject, err := filepath.Abs(projectPath)
			if err != nil {
				return "", err
			}
			if !strings.HasPrefix(cleanPath, cleanProject) {
				return "", fmt.Errorf("layout path outside project: %s", pathStr)
			}

			lContent, err := os.ReadFile(cleanPath)
			if err != nil {
				return "", fmt.Errorf("failed to read layout %s: %w", pathStr, err)
			}
			layoutTokens := tokenizer.Tokenize(string(lContent))
			
			layoutOutput, err := compileRecursive(layoutTokens, projectPath, newBlocks, mergedProps, pathStr)
			if err != nil {
				return "", err
			}
			builder.WriteString(layoutOutput)

		case tokenizer.SLOT:
			slotName, _ := parseArgs(token.Content)
			if blocks != nil {
				if val, ok := blocks[slotName]; ok {
					builder.WriteString(val)
				}
			}
			i++
		case tokenizer.PROP:
			propKey, _ := parseArgs(token.Content)
			if props != nil {
				if val, ok := props[propKey]; ok {
					builder.WriteString(val)
				}
			}
			i++
		case tokenizer.BLOCK, tokenizer.ENDBLOCK, tokenizer.ENDLAYOUT:
			i++
		default:
			i++
		}
	}
	return builder.String(), nil
}
func extractTokensUntil(tokens []tokenizer.Token, startIndex int, openType, closeType tokenizer.TokenType) ([]tokenizer.Token, int) {
	depth := 1
	var captured []tokenizer.Token
	i := startIndex
	for i < len(tokens) {
		t := tokens[i]
		if t.Type == openType {
			depth++
		} else if t.Type == closeType {
			depth--
			if depth == 0 {
				return captured, i + 1
			}
		}
		captured = append(captured, t)
		i++
	}
	return captured, i
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
			if path == filepath.Join(templatesPath, "admin") {
				return fmt.Errorf("the 'admin' directory is reserved")
			}
			if path == filepath.Join(templatesPath, "static") {
				return fmt.Errorf("the 'static' directory is reserved")
			}
			return nil
		}
		if path == filepath.Join(templatesPath, "login.html") {
			return fmt.Errorf("the 'login.html' file is reserved")
		}
		if path == filepath.Join(templatesPath, "admin.html") {
			return fmt.Errorf("the 'admin.html' file is reserved")
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
			
			compiledContent, err := Compile(tokens, projectPath, filepath.Join("templates", relativePath))
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
            "content: 'Edit / Delete';" +
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

      // 3. Inject Admin Control Buttons
      const sourcePath = document.body.getAttribute('data-source-path');
      if (sourcePath) {
          const container = document.createElement('div');
          container.style.cssText = 'position:fixed;bottom:2rem;left:2rem;z-index:9999;display:flex;gap:0.75rem;';

          const editBtn = document.createElement('a');
          editBtn.href = '/admin/files/view?path=' + encodeURIComponent(sourcePath);
          editBtn.textContent = 'Edit';
          editBtn.style.cssText = 'background:#171717;color:white;padding:0.5rem 1rem;border-radius:0.375rem;font-family:sans-serif;font-size:0.875rem;text-decoration:none;border:1px solid #404040;box-shadow:0 4px 6px -1px rgba(0,0,0,0.1);';
          editBtn.onmouseover = () => editBtn.style.background = '#262626';
          editBtn.onmouseout = () => editBtn.style.background = '#171717';
          container.appendChild(editBtn);

          const homeBtn = document.createElement('a');
          homeBtn.href = '/admin';
          homeBtn.textContent = 'Home';
          homeBtn.style.cssText = 'background:#171717;color:white;padding:0.5rem 1rem;border-radius:0.375rem;font-family:sans-serif;font-size:0.875rem;text-decoration:none;border:1px solid #404040;box-shadow:0 4px 6px -1px rgba(0,0,0,0.1);';
          homeBtn.onmouseover = () => homeBtn.style.background = '#262626';
          homeBtn.onmouseout = () => homeBtn.style.background = '#171717';
          container.appendChild(homeBtn);

          document.body.appendChild(container);
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
            
            let el = e.target;
            
            // Helper to find previous comment sibling
            function findMarker(element) {
                let ptr = element;
                while (ptr && ptr !== document.body) {
                    let balance = 0;
                    let sibling = ptr.previousSibling;
                    while (sibling) {
                        if (sibling.nodeType === 8) { // Comment
                            const text = sibling.nodeValue.trim();
                            if (text === '__TP_END_INC__') {
                                balance++;
                            } else if (text.startsWith('__TP_INC__')) {
                                if (balance === 0) {
                                    const fileMatch = text.match(/file=\"([^\"]+)\"/);
                                    const indexMatch = text.match(/token_index=\"([^\"]+)\"/);
                                    if (fileMatch && indexMatch) {
                                        return { file: fileMatch[1], index: indexMatch[1] };
                                    }
                                } else {
                                    balance--;
                                }
                            }
                        }
                        sibling = sibling.previousSibling;
                    }
                    ptr = ptr.parentNode;
                }
                return null;
            }
            
            const info = findMarker(el);
            
            if (info) {
                if (confirm('Delete this block included from ' + info.file + '?')) {
                    fetch('/admin/api/delete-block', {
                        method: 'POST',
                        headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                        body: 'file=' + encodeURIComponent(info.file) + '&token_index=' + encodeURIComponent(info.index)
                    }).then(res => {
                        if (res.ok) {
                            window.location.reload();
                        } else {
                            alert('Error deleting block');
                        }
                    });
                }
            } else {
                alert("Could not identify the source of this block. It might be hardcoded in the layout.");
            }
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
