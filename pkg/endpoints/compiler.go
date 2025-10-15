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
	CacheDir string
	BaseDir  string
	cache    map[string]*cacheEntry
}

type cacheEntry struct {
	endpoint *LoadedEndpoint
	modTime  time.Time
	soPath   string
}

func NewCompiler(baseDir, cacheDir string) *EndpointCompiler {
	return &EndpointCompiler{
		CacheDir: cacheDir,
		BaseDir:  baseDir,
		cache:    make(map[string]*cacheEntry),
	}
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
		return nil, fmt.Errorf("no HTTP method handlers found")
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
		return nil
	}

	src := string(content)
	methods := []string{}
	httpMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "ALL"}

	for _, method := range httpMethods {
		pattern := fmt.Sprintf("func %s(", method)
		if strings.Contains(src, pattern) {
			methods = append(methods, method)
		}
	}

	return methods
}

func (c *EndpointCompiler) compile(filePath string, methods []string) (string, error) {
	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return "", err
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))[:8]
	soPath := filepath.Join(c.CacheDir, fmt.Sprintf("endpoint-%s.so", hash))

	pkgAlias := strings.TrimSuffix(filepath.Base(filePath), ".go")
	relPath, _ := filepath.Rel(c.BaseDir, filepath.Dir(filePath))
	importPath := filepath.Join(filepath.Base(c.BaseDir), relPath)

	var methodExports strings.Builder
	for _, method := range methods {
		methodExports.WriteString(fmt.Sprintf("var %s = %s.%s\n", method, pkgAlias, method))
	}

	pluginSrc := fmt.Sprintf(`package main

import (
	"github.com/galaxy/galaxy/pkg/endpoints"
	%s "%s"
)

%s
`, pkgAlias, importPath, methodExports.String())

	pluginPath := filepath.Join(c.CacheDir, fmt.Sprintf("endpoint-%s.go", hash))
	if err := os.WriteFile(pluginPath, []byte(pluginSrc), 0644); err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soPath, pluginPath)
	cmd.Dir = c.BaseDir
	output, err := cmd.CombinedOutput()
	if err != nil {
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
