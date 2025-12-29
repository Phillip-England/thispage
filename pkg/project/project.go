package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// New creates a new directory with the given name, and "live", "partials", and "templates" subdirectories inside it.
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
	liveDirPath := filepath.Join(name, "live")
	partialsDirPath := filepath.Join(name, "partials")
	templatesDirPath := filepath.Join(name, "templates")

	// Create the "live" subdirectory
	err = os.Mkdir(liveDirPath, 0755)
	if err != nil {
		_ = os.RemoveAll(name) // Clean up parent directory
		return fmt.Errorf("could not create 'live' subdirectory in '%s': %w", name, err)
	}

	// Create the "partials" subdirectory
	err = os.Mkdir(partialsDirPath, 0755)
	if err != nil {
		_ = os.RemoveAll(name) // Clean up parent directory
		return fmt.Errorf("could not create 'partials' subdirectory in '%s': %w", name, err)
	}

	// Create the "templates" subdirectory
	err = os.Mkdir(templatesDirPath, 0755)
	if err != nil {
		_ = os.RemoveAll(name) // Clean up parent directory
		return fmt.Errorf("could not create 'templates' subdirectory in '%s': %w", name, err)
	}

	// Create default files and directories
	postsTemplatesPath := filepath.Join(templatesDirPath, "posts")
	if err := os.Mkdir(postsTemplatesPath, 0755); err != nil {
		_ = os.RemoveAll(name)
		return fmt.Errorf("could not create 'posts' subdirectory in 'templates': %w", err)
	}

	defaultIndexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Home - This Page</title>
</head>
<body>
  <h1>Home Page</h1>
  {{ include ./partials/navigation.html }}
</body>
</html>`

	defaultAboutHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>About - This Page</title>
</head>
<body>
  <h1>About Page</h1>
  {{ include ./partials/navigation.html }}
</body>
</html>`

	defaultPostHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>First Post - This Page</title>
</head>
<body>
  <h1>First Post</h1>
  {{ include ./partials/navigation.html }}
</body>
</html>`

	defaultNavigationHTML := `<nav>
  <ul>
    <li>
      <a href='/'>Home</a>
    </li>
    <li>
      <a href='/about'>About</a>
    </li>
    <li>
      <a href='/posts/1'>First Post</a>
    </li>
  </ul>
</nav>`

	filesToCreate := map[string]string{
		filepath.Join(templatesDirPath, "index.html"):     defaultIndexHTML,
		filepath.Join(templatesDirPath, "about.html"):     defaultAboutHTML,
		filepath.Join(postsTemplatesPath, "1.html"): defaultPostHTML,
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
