package templates

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed minimal blog portfolio documentation
var templateFS embed.FS

type TemplateData struct {
	ProjectName string
}

type Generator struct {
	templateFS   embed.FS
	templateName string
	data         TemplateData
}

func NewGenerator(templateName string, projectName string) (*Generator, error) {
	validTemplates := map[string]bool{
		"minimal":       true,
		"blog":          true,
		"portfolio":     true,
		"documentation": true,
	}

	if !validTemplates[templateName] {
		return nil, fmt.Errorf("unknown template: %s", templateName)
	}

	return &Generator{
		templateFS:   templateFS,
		templateName: templateName,
		data:         TemplateData{ProjectName: projectName},
	}, nil
}

func (g *Generator) Generate(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	return g.copyDir(g.templateName, targetDir)
}

func (g *Generator) copyDir(srcDir, dstDir string) error {
	entries, err := g.templateFS.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := g.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := g.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) copyFile(srcPath, dstPath string) error {
	content, err := g.templateFS.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if strings.HasSuffix(srcPath, ".toml") || strings.HasSuffix(srcPath, ".md") {
		tmpl, err := template.New("file").Parse(string(content))
		if err != nil {
			return err
		}

		f, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer f.Close()

		return tmpl.Execute(f, g.data)
	}

	return os.WriteFile(dstPath, content, 0644)
}

func InstallDependencies(projectDir, packageManager string) error {
	cmd := exec.Command(packageManager, "install")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func InitGit(projectDir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
