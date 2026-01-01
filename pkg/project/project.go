package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// New creates a new directory with the given name, and standard subdirectories and files.
func New(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Create the main project directory
	err := os.Mkdir(name, 0755)
	if err != nil {
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
	postsTemplatesPath := filepath.Join(templatesDirPath, "posts")

	defaultIndexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>Home - This Page</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen">
  <div class="max-w-2xl mx-auto py-20 px-6">
    {{ include ./partials/navigation.html }}
    <main class="mt-12">
        <h1 class="text-4xl font-bold text-white tracking-tight">Home Page</h1>
        <p class="mt-4 text-neutral-500 leading-relaxed">
            Welcome to your new thispage project. This is the index page. 
            You can find this file at <code class="bg-neutral-900 px-1 rounded text-blue-400">templates/index.html</code>.
        </p>
    </main>
  </div>
</body>
</html>`

	defaultAboutHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>About - This Page</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen">
  <div class="max-w-2xl mx-auto py-20 px-6">
    {{ include ./partials/navigation.html }}
    <main class="mt-12">
        <h1 class="text-4xl font-bold text-white tracking-tight">About Page</h1>
        <p class="mt-4 text-neutral-500 leading-relaxed">
            This is the about page. You can customize this in <code class="bg-neutral-900 px-1 rounded text-blue-400">templates/about.html</code>.
        </p>
    </main>
  </div>
</body>
</html>`

	defaultPostHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <script src="https://cdn.tailwindcss.com"></script>
  <title>First Post - This Page</title>
</head>
<body class="bg-neutral-950 text-neutral-300 font-sans antialiased min-h-screen">
  <div class="max-w-2xl mx-auto py-20 px-6">
    {{ include ./partials/navigation.html }}
    <main class="mt-12">
        <h1 class="text-4xl font-bold text-white tracking-tight">First Post</h1>
        <p class="mt-4 text-neutral-500 leading-relaxed">
            This is your first blog post. Find it at <code class="bg-neutral-900 px-1 rounded text-blue-400">templates/posts/1.html</code>.
        </p>
    </main>
  </div>
</body>
</html>`

	defaultNavigationHTML := `<nav class="flex gap-6 border-b border-neutral-800 pb-6">
  <a href='/' class="text-xs uppercase tracking-widest hover:text-white transition-colors">Home</a>
  <a href='/about' class="text-xs uppercase tracking-widest hover:text-white transition-colors">About</a>
  <a href='/posts/1' class="text-xs uppercase tracking-widest hover:text-white transition-colors">First Post</a>
</nav>`

	filesToCreate := map[string]string{
		filepath.Join(templatesDirPath, "index.html"):     defaultIndexHTML,
		filepath.Join(templatesDirPath, "about.html"):     defaultAboutHTML,
		filepath.Join(postsTemplatesPath, "1.html"):        defaultPostHTML,
		filepath.Join(partialsDirPath, "navigation.html"): defaultNavigationHTML,
	}

	for path, content := range filesToCreate {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			_ = os.RemoveAll(name)
			return fmt.Errorf("could not create default file '%s': %w", filepath.Base(path), err)
		}
	}

	return nil
}