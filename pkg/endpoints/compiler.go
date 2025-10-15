package endpoints

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type EndpointCompiler struct {
	CacheDir   string
	BaseDir    string
	ModuleName string
	cache      map[string]*cacheEntry
}

type cacheEntry struct {
	endpoint *LoadedEndpoint
	modTime  time.Time
	soPath   string
}

func NewCompiler(baseDir, cacheDir string) *EndpointCompiler {
	moduleName := detectModuleName(baseDir)
	return &EndpointCompiler{
		CacheDir:   cacheDir,
		BaseDir:    baseDir,
		ModuleName: moduleName,
		cache:      make(map[string]*cacheEntry),
	}
}

func detectModuleName(baseDir string) string {
	modPath := filepath.Join(baseDir, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return filepath.Base(baseDir)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return filepath.Base(baseDir)
}

func (c *EndpointCompiler) Load(filePath string) (*LoadedEndpoint, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	if entry, ok := c.cache[filePath]; ok {
		if entry.modTime.Equal(info.ModTime()) {
			return entry.endpoint, nil
		}
	}

	methods := c.detectMethods(filePath)
	if len(methods) == 0 {
		return nil, fmt.Errorf("no HTTP method handlers found in %s", filePath)
	}

	soPath, err := c.compile(filePath, methods)
	if err != nil {
		return nil, err
	}

	endpoint, err := LoadPlugin(soPath)
	if err != nil {
		return nil, err
	}

	c.cache[filePath] = &cacheEntry{
		endpoint: endpoint,
		modTime:  info.ModTime(),
		soPath:   soPath,
	}

	return endpoint, nil
}

func (c *EndpointCompiler) detectMethods(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("DEBUG: Failed to read %s: %v\n", filePath, err)
		return nil
	}

	src := string(content)
	methods := []string{}
	httpMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "ALL"}

	for _, method := range httpMethods {
		pattern := fmt.Sprintf("func %s(", method)
		if strings.Contains(src, pattern) {
			methods = append(methods, method)
			fmt.Printf("DEBUG: Found method %s in %s\n", method, filePath)
		}
	}

	fmt.Printf("DEBUG: Detected %d methods in %s\n", len(methods), filePath)
	return methods
}

func (c *EndpointCompiler) compile(filePath string, methods []string) (string, error) {
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))[:8]
	soPath := filepath.Join(c.CacheDir, fmt.Sprintf("endpoint-%s.so", hash))

	dir := filepath.Dir(filePath)
	pkgName := filepath.Base(dir)
	relPath, _ := filepath.Rel(c.BaseDir, dir)
	importPath := filepath.Join(c.ModuleName, relPath)

	var methodExports strings.Builder
	for _, method := range methods {
		methodExports.WriteString(fmt.Sprintf("func %s(ctx *endpoints.Context) error { return %s.%s(ctx) }\n", method, pkgName, method))
	}

	pluginSrc := fmt.Sprintf(`package main

import (
	"github.com/galaxy/galaxy/pkg/endpoints"
	%s "%s"
)

%s
`, pkgName, importPath, methodExports.String())

	pluginPath := filepath.Join(c.CacheDir, fmt.Sprintf("endpoint-%s.go", hash))
	if err := os.WriteFile(pluginPath, []byte(pluginSrc), 0644); err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: Generated plugin:\n%s\n", pluginSrc)

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soPath, pluginPath)
	cmd.Dir = c.BaseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DEBUG: Compile failed for %s\n", pluginPath)
		fmt.Printf("DEBUG: Command: go build -buildmode=plugin -o %s %s\n", soPath, pluginPath)
		fmt.Printf("DEBUG: Working dir: %s\n", c.BaseDir)
		fmt.Printf("DEBUG: Output: %s\n", string(output))
		return "", fmt.Errorf("compile plugin: %w\n%s", err, string(output))
	}

	return soPath, nil
}

func (c *EndpointCompiler) getPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	base := filepath.Base(dir)
	if base == "api" || base == "." {
		return "api"
	}
	return base
}
