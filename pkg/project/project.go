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
	dirs := []string{"live", "partials", "templates", "templates/posts", "templates/static"}

	for _, dir := range dirs {
		dirPath := filepath.Join(name, dir)
		err = os.MkdirAll(dirPath, 0755)
		if err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create subdirectory '%s': %w", dir, err)
		}
	}

	templatesDirPath := filepath.Join(name, "templates")
	partialsDirPath := filepath.Join(name, "partials")

	defaultIndexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>Home - This Page</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen relative">
  <div class="max-w-4xl mx-auto py-20 px-6">
    {{ include ./partials/navigation.html }}
    
    <main class="mt-12 min-h-[50vh] relative">
        <!-- Empty State Message -->
        <div class="empty-state absolute inset-0 flex items-center justify-center border-2 border-dashed border-neutral-800 rounded-lg">
            <a href="/admin" class="text-neutral-700 hover:text-neutral-500 text-sm font-mono transition-colors">
                Login to start building
            </a>
        </div>
    </main>
  </div>
  
  <!-- The Add Content Button (Visible only in Admin Mode) -->
  <div class="thispage-add-btn hidden fixed bottom-8 right-8 z-50 cursor-pointer hover:scale-110 transition-transform group">
        <div class="w-14 h-14 rounded-full bg-blue-600 flex items-center justify-center text-white text-3xl font-bold shadow-lg hover:bg-blue-500 transition-colors">
            +
        </div>
        <div class="absolute right-full mr-4 top-1/2 -translate-y-1/2 bg-neutral-900 text-white text-xs px-3 py-1 rounded opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">
            Add Block
        </div>
  </div>
</body>
</html>`

	defaultNavigationHTML := `<nav class="flex gap-6 border-b border-neutral-800 pb-6 w-full mb-8">
  <a href='/' class="text-xs uppercase tracking-widest hover:text-white transition-colors">Home</a>
</nav>`

    defaultHeroHTML := `<header class="py-20 text-center thispage-block">
    <h1 class="text-5xl font-bold text-white tracking-tight mb-6">Welcome to Your Site</h1>
    <p class="text-xl text-neutral-400 max-w-2xl mx-auto">This is a hero section. It's a great place to introduce your brand or project.</p>
</header>`

    defaultContentHTML := `<section class="py-12 thispage-block">
    <h2 class="text-2xl font-bold text-white mb-4">Content Section</h2>
    <p class="text-neutral-400 leading-relaxed">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
        Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
    </p>
</section>`

    defaultFooterHTML := `<footer class="border-t border-neutral-800 py-8 mt-12 text-center text-sm text-neutral-600 thispage-block">
    &copy; 2024 Your Company. All rights reserved.
</footer>`

	filesToCreate := map[string]string{
		filepath.Join(templatesDirPath, "index.html"):         defaultIndexHTML,
		filepath.Join(partialsDirPath, "navigation.html"):     defaultNavigationHTML,
		filepath.Join(partialsDirPath, "hero.html"):           defaultHeroHTML,
		filepath.Join(partialsDirPath, "content.html"):        defaultContentHTML,
		filepath.Join(partialsDirPath, "footer.html"):          defaultFooterHTML,
		filepath.Join(name, "templates/static/input.css"): "@import \"tailwindcss\";\n",
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create default file '%s': %w", filepath.Base(path), err)
		}
	}

	return nil
}
