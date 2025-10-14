package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gastro/gastro/internal/assets"
	"github.com/gastro/gastro/pkg/executor"
	"github.com/gastro/gastro/pkg/parser"
	"github.com/gastro/gastro/pkg/template"
)

type ComponentCompiler struct {
	BaseDir string
	Cache   map[string]*parser.Component
	Bundler *assets.Bundler
}

func NewComponentCompiler(baseDir string) *ComponentCompiler {
	return &ComponentCompiler{
		BaseDir: baseDir,
		Cache:   make(map[string]*parser.Component),
		Bundler: assets.NewBundler(".tmp"),
	}
}

func (c *ComponentCompiler) Compile(filePath string, props map[string]interface{}, slots map[string]string) (string, error) {
	comp, err := c.loadComponent(filePath)
	if err != nil {
		return "", err
	}

	ctx := executor.NewContext()
	for k, v := range props {
		ctx.SetProp(k, v)
		ctx.Set(k, v)
	}

	if comp.Frontmatter != "" {
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			return "", err
		}
	}

	processedTemplate := c.ProcessComponentTags(comp.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, &template.RenderOptions{
		Props: props,
		Slots: slots,
	})
	if err != nil {
		return "", err
	}

	return rendered, nil
}

func (c *ComponentCompiler) loadComponent(filePath string) (*parser.Component, error) {
	if comp, ok := c.Cache[filePath]; ok {
		return comp, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		return nil, err
	}

	c.Cache[filePath] = comp
	return comp, nil
}

var (
	componentOpenCloseRegex = regexp.MustCompile(`(?s)<([A-Z]\w+)([^>]*)>(.*?)</([A-Z]\w+)>`)
	componentSelfCloseRegex = regexp.MustCompile(`<([A-Z]\w+)([^/>]*)/?>`)
)

func (c *ComponentCompiler) ProcessComponentTags(template string, ctx *executor.Context) string {
	result := componentOpenCloseRegex.ReplaceAllStringFunc(template, func(match string) string {
		matches := componentOpenCloseRegex.FindStringSubmatch(match)

		componentName := matches[1]
		attrs := matches[2]
		content := matches[3]
		closingTag := matches[4]

		if componentName != closingTag {
			return match
		}

		componentPath := filepath.Join(c.BaseDir, "components", componentName+".gxc")

		props := c.parseAttributes(attrs, ctx)

		slots := make(map[string]string)
		trimmedContent := strings.TrimSpace(content)
		if trimmedContent != "" {
			slots["default"] = trimmedContent
		}

		rendered, err := c.Compile(componentPath, props, slots)
		if err != nil {
			return fmt.Sprintf("<!-- Error rendering %s: %v -->", componentName, err)
		}

		return rendered
	})

	result = componentSelfCloseRegex.ReplaceAllStringFunc(result, func(match string) string {
		matches := componentSelfCloseRegex.FindStringSubmatch(match)

		componentName := matches[1]
		attrs := matches[2]

		componentPath := filepath.Join(c.BaseDir, "components", componentName+".gxc")

		props := c.parseAttributes(attrs, ctx)

		rendered, err := c.Compile(componentPath, props, make(map[string]string))
		if err != nil {
			return fmt.Sprintf("<!-- Error rendering %s: %v -->", componentName, err)
		}

		return rendered
	})

	return result
}

func (c *ComponentCompiler) parseAttributes(attrs string, ctx *executor.Context) map[string]interface{} {
	props := make(map[string]interface{})

	attrRegex := regexp.MustCompile(`(\w+)=\{([^}]+)\}|(\w+)="([^"]+)"|(\w+)='([^']+)'`)
	matches := attrRegex.FindAllStringSubmatch(attrs, -1)

	for _, match := range matches {
		if match[1] != "" {
			varName := match[2]
			if val, ok := ctx.Get(varName); ok {
				props[match[1]] = val
			} else {
				props[match[1]] = varName
			}
		} else if match[3] != "" {
			props[match[3]] = match[4]
		} else if match[5] != "" {
			props[match[5]] = match[6]
		}
	}

	return props
}
