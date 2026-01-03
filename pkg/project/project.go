package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// New creates a new directory with the given name, and standard subdirectories and files.
func New(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

    if force {
        if err := os.RemoveAll(name); err != nil {
             return fmt.Errorf("failed to remove existing directory '%s': %w", name, err)
        }
    }

	// Create the main project directory
	err := os.Mkdir(name, 0755)
	if err != nil {
        if os.IsExist(err) {
             return fmt.Errorf("directory '%s' already exists. Use --force to overwrite", name)
        }
		return fmt.Errorf("could not create project directory '%s': %w", name, err)
	}

	// Define subdirectory paths
	dirs := []string{"live", "components", "containers", "templates", "templates/posts", "static", "layouts", "content"}

	for _, dir := range dirs {
		dirPath := filepath.Join(name, dir)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create subdirectory '%s': %w", dir, err)
		}
	}

	templatesDirPath := filepath.Join(name, "templates")
	componentsDirPath := filepath.Join(name, "components")
	layoutsDirPath := filepath.Join(name, "layouts")
	contentDirPath := filepath.Join(name, "content")

    guestLayoutHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>{{ slot "title" }}</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen relative">
  <div class="max-w-4xl mx-auto py-20 px-6">
    {{ include "./components/navigation.html" }}
    
    <main class="mt-12 min-h-[50vh] relative">
        {{ slot "main" }}
    </main>
  </div>
</body>
</html>`

	defaultIndexHTML := `{{ layout "./layouts/guest_layout.html" }}

{{ block "title" }}Home - This Page{{ endblock }}

{{ block "main" }}
    <div class="flex flex-col items-center justify-center min-h-[50vh] text-center">
        <h1 class="text-4xl font-bold text-white mb-4">Welcome</h1>
        <p class="text-neutral-400 mb-8">Login to start building your site.</p>
        <a href="/login" class="bg-blue-600 text-white px-6 py-3 rounded hover:bg-blue-500 transition-colors">Login</a>
    </div>
{{ endblock }}

{{ endlayout }}`

	defaultNavigationHTML := `<nav class="flex gap-6 border-b border-neutral-800 pb-6 w-full mb-8 items-center thispage-component">
  <a href='/' class="font-bold text-white text-lg mr-auto">ThisPage</a>
</nav>`

    defaultFooterHTML := `<footer class="border-t border-neutral-800 py-8 mt-12 text-center text-sm text-neutral-600 thispage-component">
    &copy; 2024 Your Company.
</footer>`

	defaultWelcomeMD := `# Welcome to ThisPage

This is your first markdown file. You can use this directory to store content for your site.

## Getting Started

- Edit this file to add your own content
- Create new markdown files for blog posts, documentation, etc.
- Use the file manager to organize your content
`

	filesToCreate := map[string]string{
		filepath.Join(templatesDirPath, "index.html"):       defaultIndexHTML,
		filepath.Join(layoutsDirPath, "guest_layout.html"):  guestLayoutHTML,
		filepath.Join(componentsDirPath, "navigation.html"): defaultNavigationHTML,
		filepath.Join(componentsDirPath, "footer.html"):     defaultFooterHTML,
		filepath.Join(contentDirPath, "welcome.md"):         defaultWelcomeMD,
		filepath.Join(name, "static/input.css"):             "@import \"tailwindcss\";\n",
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create default file '%s': %w", filepath.Base(path), err)
		}
	}

	return nil
}
