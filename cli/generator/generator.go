package generator

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

//go:embed project_template/*.gotpl project_template/*/*/*.gotpl
var templates embed.FS

type ProjectData struct {
	Name       string // Project name
	ModulePath string // Full module path (e.g., github.com/username/project)
}

func downloadTailwindCSS(projectDir string) error {
	// Use Tailwind's CDN URL for the macOS binary
	url := "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64"

	fmt.Printf("Downloading Tailwind CSS from: %s\n", url)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Allow redirects
		},
	}

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent header to avoid potential API restrictions
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/octet-stream")

	// Download the file
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download Tailwind CSS: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to download Tailwind CSS (HTTP %d): %s\nURL: %s", resp.StatusCode, string(body), url)
	}

	// Create the output file
	outputPath := filepath.Join(projectDir, "tailwindcss")

	// Create output file
	outFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Copy the response body directly to the file
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		os.Remove(outputPath) // Clean up on error
		return fmt.Errorf("failed to write binary: %w", err)
	}

	return nil
}

func GenerateProject(name string) error {
	data := ProjectData{
		Name:       name,
		ModulePath: fmt.Sprintf("github.com/brainmaniac/%s", name),
	}

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create necessary directories
	dirsToCreate := []string{
		filepath.Join(name, "cmd", "server"),
		filepath.Join(name, "internal", "app"),
		filepath.Join(name, "internal", "handlers"),
		filepath.Join(name, "views", "layouts"),
		filepath.Join(name, "views", "pages"),
		filepath.Join(name, "assets", "css"),
		filepath.Join(name, "public", "css"),
		filepath.Join(name, "tmp"),
	}

	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Initialize the module first
	cmd := exec.Command("go", "mod", "init", data.ModulePath)
	cmd.Dir = name
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to init go mod: %v\nOutput: %s", err, string(out))
	}

	// Process all templates
	entries, err := templates.ReadDir("project_template")
	if err != nil {
		return fmt.Errorf("failed to read templates: %w", err)
	}

	for _, entry := range entries {
		if err := processEntry("project_template", name, entry, data); err != nil {
			return fmt.Errorf("failed to process %s: %w", entry.Name(), err)
		}
	}

	// Download Tailwind CSS
	fmt.Println("📦 Downloading Tailwind CSS...")
	if err := downloadTailwindCSS(name); err != nil {
		return fmt.Errorf("failed to download Tailwind CSS: %w", err)
	}

	// Run templ generate
	cmd = exec.Command("templ", "generate")
	cmd.Dir = name
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate templ files: %v\nOutput: %s", err, string(out))
	}

	// Then run go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = name
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %v\nOutput: %s", err, string(out))
	}

	fmt.Println("🤘🎸 Project created successfully!")
	fmt.Println("The initial CSS will be built when you run the project.")
	return nil
}

func processEntry(base, dest string, entry os.DirEntry, data ProjectData) error {
	sourcePath := filepath.Join(base, entry.Name())
	destPath := filepath.Join(dest, entry.Name())

	if entry.IsDir() {
		// Create the directory
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return err
		}
		// Process contents of the directory
		entries, err := templates.ReadDir(sourcePath)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if err := processEntry(sourcePath, destPath, e, data); err != nil {
				return err
			}
		}
		return nil
	}

	return processTemplateFile(sourcePath, destPath, data)
}

func processTemplateFile(src, dest string, data ProjectData) error {
	content, err := templates.ReadFile(src)
	if err != nil {
		return err
	}

	// Handle .gotpl extension removal
	destPath := dest
	if strings.HasSuffix(dest, ".gotpl") {
		destPath = dest[:len(dest)-6] // remove .gotpl suffix
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Create the destination file
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Parse and execute the template
	tmpl, err := template.New(filepath.Base(src)).Parse(string(content))
	if err != nil {
		return err
	}

	return tmpl.Execute(f, data)
}

func initGoMod(projectPath, modulePath string) error {
	cmd := exec.Command("go", "mod", "init", modulePath)
	cmd.Dir = projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to init go mod: %v\nCommand output: %s", err, string(output))
	}

	return nil
}
