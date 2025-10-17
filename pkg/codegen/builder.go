package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cameron-webmatter/galaxy/pkg/parser"
	"github.com/cameron-webmatter/galaxy/pkg/router"
)

type CodegenBuilder struct {
	Routes         []*router.Route
	PagesDir       string
	OutDir         string
	ModuleName     string
	MiddlewarePath string
}

func NewCodegenBuilder(routes []*router.Route, pagesDir, outDir, moduleName string) *CodegenBuilder {
	srcDir := filepath.Dir(pagesDir)
	middlewarePath := filepath.Join(srcDir, "middleware.go")

	return &CodegenBuilder{
		Routes:         routes,
		PagesDir:       pagesDir,
		OutDir:         outDir,
		ModuleName:     moduleName,
		MiddlewarePath: middlewarePath,
	}
}

func (b *CodegenBuilder) Build() error {
	serverDir := filepath.Join(b.OutDir, "server")

	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return err
	}

	var handlers []*GeneratedHandler
	var nonEndpointRoutes []*router.Route

	for _, route := range b.Routes {
		if route.IsEndpoint {
			continue
		}

		content, err := os.ReadFile(route.FilePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", route.FilePath, err)
		}

		comp, err := parser.Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse %s: %w", route.FilePath, err)
		}

		gen := NewHandlerGenerator(comp, route, b.ModuleName, b.PagesDir)
		handler, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("generate handler for %s: %w", route.Pattern, err)
		}

		handlers = append(handlers, handler)
		nonEndpointRoutes = append(nonEndpointRoutes, route)
	}

	manifestPath := filepath.Join(serverDir, "_assets", "wasm-manifest.json")
	hasMiddleware := false
	if _, err := os.Stat(b.MiddlewarePath); err == nil {
		hasMiddleware = true
	}

	mainGen := NewMainGenerator(handlers, nonEndpointRoutes, b.ModuleName, manifestPath)
	mainGen.HasMiddleware = hasMiddleware
	mainGo := mainGen.Generate()

	if err := os.WriteFile(filepath.Join(serverDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return err
	}

	runtimeDir := filepath.Join(serverDir, "runtime")
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return err
	}

	runtime := mainGen.GenerateRuntime()
	if err := os.WriteFile(filepath.Join(runtimeDir, "runtime.go"), []byte(runtime), 0644); err != nil {
		return err
	}

	if err := b.generateGoMod(serverDir); err != nil {
		return err
	}

	if err := b.copyMiddleware(serverDir); err != nil {
		return fmt.Errorf("copy middleware: %w", err)
	}

	if err := b.compile(serverDir); err != nil {
		return err
	}

	return nil
}

func (b *CodegenBuilder) copyMiddleware(serverDir string) error {
	if _, err := os.Stat(b.MiddlewarePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(b.MiddlewarePath)
	if err != nil {
		return err
	}

	content := string(data)
	content = regexp.MustCompile(`(?m)^package\s+\w+`).ReplaceAllString(content, "package main")

	destPath := filepath.Join(serverDir, "middleware.go")
	return os.WriteFile(destPath, []byte(content), 0644)
}

func (b *CodegenBuilder) generateGoMod(serverDir string) error {
	galaxyPath, err := findGalaxyRoot()
	if err != nil {
		galaxyPath = "../../.."
	}

	goMod := fmt.Sprintf(`module %s

go 1.23

replace github.com/cameron-webmatter/galaxy => %s

require github.com/cameron-webmatter/galaxy v0.0.0
`, b.ModuleName, galaxyPath)

	return os.WriteFile(filepath.Join(serverDir, "go.mod"), []byte(goMod), 0644)
}

func findGalaxyRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for dir := wd; dir != "/"; dir = filepath.Dir(dir) {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err != nil {
			continue
		}

		if strings.Contains(string(data), "module github.com/cameron-webmatter/galaxy") {
			return dir, nil
		}
	}

	return "", fmt.Errorf("galaxy root not found")
}

func (b *CodegenBuilder) compile(serverDir string) error {
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = serverDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, output)
	}

	cmd := exec.Command("go", "build", "-o", "server", ".")
	cmd.Dir = serverDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compile server: %w\n%s", err, output)
	}
	return nil
}
