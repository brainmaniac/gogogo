// cli/generator/generator.go
package generator

import (
	"embed"
	"fmt"
	"os"
	"os/exec" // add this for exec.Command
	"path/filepath"
	"text/template"
)

//go:embed project_template
var templates embed.FS

type ProjectData struct {
	Name       string // Project name
	ModulePath string // Full module path (e.g., github.com/username/project)
}

func GenerateProject(name string) error {
	data := ProjectData{
		Name:       name,
		ModulePath: fmt.Sprintf("github.com/yourusername/%s", name),
	}

	// Create project directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Walk through the embedded templates
	entries, err := templates.ReadDir("project_template")
	if err != nil {
		return fmt.Errorf("failed to read templates: %w", err)
	}

	for _, entry := range entries {
		if err := processEntry("project_template", name, entry, data); err != nil {
			return fmt.Errorf("failed to process %s: %w", entry.Name(), err)
		}
	}

	// Initialize go.mod
	if err := initGoMod(name, data.ModulePath); err != nil {
		return fmt.Errorf("failed to initialize go.mod: %w", err)
	}

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

	// Process file
	return processTemplateFile(sourcePath, destPath, data)
}

func processTemplateFile(src, dest string, data ProjectData) error {
	content, err := templates.ReadFile(src)
	if err != nil {
		return err
	}

	// If it's a .gotpl file, remove the extension
	if filepath.Ext(dest) == ".gotpl" {
		dest = dest[:len(dest)-6] // remove .gotpl
	}

	// Create the destination file
	f, err := os.Create(dest)
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
	return cmd.Run()
}
