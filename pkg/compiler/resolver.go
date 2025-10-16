package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ComponentResolver struct {
	BaseDir        string
	ComponentDirs  []string
	CurrentFile    string
	ExplicitPaths  map[string]string
	Cache          map[string]string
	ComponentIndex map[string]string
}

func NewComponentResolver(baseDir string, componentDirs []string) *ComponentResolver {
	if componentDirs == nil {
		componentDirs = []string{"components"}
	}

	resolver := &ComponentResolver{
		BaseDir:        baseDir,
		ComponentDirs:  componentDirs,
		ExplicitPaths:  make(map[string]string),
		Cache:          make(map[string]string),
		ComponentIndex: make(map[string]string),
	}

	resolver.buildComponentIndex()
	return resolver
}

func (r *ComponentResolver) buildComponentIndex() {
	filepath.Walk(r.BaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			dirName := filepath.Base(path)
			if dirName == "pages" || dirName == "node_modules" || dirName == ".git" || strings.HasPrefix(dirName, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) == ".gxc" {
			baseName := strings.TrimSuffix(filepath.Base(path), ".gxc")
			if _, exists := r.ComponentIndex[baseName]; !exists {
				r.ComponentIndex[baseName] = path
			}
		}

		return nil
	})
}

func (r *ComponentResolver) SetCurrentFile(file string) {
	r.CurrentFile = file
}

func (r *ComponentResolver) ParseImports(imports []Import) {
	for _, imp := range imports {
		if imp.IsComponent {
			r.ExplicitPaths[imp.Alias] = imp.Path
		}
	}
}

type Import struct {
	Path        string
	Alias       string
	IsComponent bool
}

func (r *ComponentResolver) ExtractComponentRefs(template string) []string {
	componentNameRegex := regexp.MustCompile(`<([A-Z]\w+)`)
	matches := componentNameRegex.FindAllStringSubmatch(template, -1)

	seen := make(map[string]bool)
	var refs []string

	for _, match := range matches {
		name := match[1]
		if !seen[name] {
			refs = append(refs, name)
			seen[name] = true
		}
	}

	return refs
}

func (r *ComponentResolver) Resolve(name string) (string, error) {
	if cached, ok := r.Cache[name]; ok {
		return cached, nil
	}

	if path, ok := r.ExplicitPaths[name]; ok {
		resolved, err := r.resolveImportPath(path)
		if err != nil {
			return "", err
		}
		r.Cache[name] = resolved
		return resolved, nil
	}

	if path, ok := r.ComponentIndex[name]; ok {
		r.Cache[name] = path
		return path, nil
	}

	if r.CurrentFile != "" {
		path := filepath.Join(filepath.Dir(r.CurrentFile), name+".gxc")
		if fileExists(path) {
			r.Cache[name] = path
			return path, nil
		}
	}

	return "", fmt.Errorf("component %s not found in %s", name, r.BaseDir)
}

func (r *ComponentResolver) resolveImportPath(importPath string) (string, error) {
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		if r.CurrentFile == "" {
			return "", fmt.Errorf("relative import requires current file context")
		}
		resolved := filepath.Join(filepath.Dir(r.CurrentFile), importPath)
		if fileExists(resolved) {
			return resolved, nil
		}
		return "", fmt.Errorf("import path not found: %s", importPath)
	}

	if strings.HasPrefix(importPath, "@/") {
		resolved := filepath.Join(r.BaseDir, strings.TrimPrefix(importPath, "@/"))
		if fileExists(resolved) {
			return resolved, nil
		}
		return "", fmt.Errorf("import path not found: %s", importPath)
	}

	resolved := filepath.Join(r.BaseDir, importPath)
	if fileExists(resolved) {
		return resolved, nil
	}

	return "", fmt.Errorf("import path not found: %s", importPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
