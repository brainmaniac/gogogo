package generator

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed project_template/*
var templates embed.FS

type ProjectData struct {
	Name       string
	ModulePath string
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

	// Copy all template files
	if err := filepath.Walk("templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("templates", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(name, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return processTemplateFile(path, destPath, data)
	}); err != nil {
		return err
	}

	// Initialize go.mod
	if err := initGoMod(name, data.ModulePath); err != nil {
		return err
	}

	return nil
}

// cli/generator/generator.go
func processTemplateFile(src, dest string, data ProjectData) error {
	content, err := templates.ReadFile(src)
	if err != nil {
		return err
	}

	// First replace our module path placeholder
	contentStr := strings.ReplaceAll(string(content), "GOGOGO_MODULE_PATH", data.ModulePath)

	// Then process any other template variables
	tmpl, err := template.New(filepath.Base(src)).Parse(contentStr)
	if err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func initGoMod(projectPath, modulePath string) error {
	cmd := exec.Command("go", "mod", "init", modulePath)
	cmd.Dir = projectPath
	return cmd.Run()
}
