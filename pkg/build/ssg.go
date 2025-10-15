package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/galaxy/galaxy/internal/assets"
	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/config"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/plugins"
	"github.com/galaxy/galaxy/pkg/plugins/tailwind"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/template"
)

type SSGBuilder struct {
	Config        *config.Config
	SrcDir        string
	PagesDir      string
	OutDir        string
	PublicDir     string
	Router        *router.Router
	Bundler       *assets.Bundler
	Compiler      *compiler.ComponentCompiler
	PluginManager *plugins.Manager
}

func NewSSGBuilder(cfg *config.Config, srcDir, pagesDir, outDir, publicDir string) *SSGBuilder {
	baseDir := srcDir

	pluginMgr := plugins.NewManager(cfg)
	pluginMgr.Register(tailwind.New())

	bundler := assets.NewBundler(outDir)
	bundler.PluginManager = pluginMgr

	return &SSGBuilder{
		Config:        cfg,
		SrcDir:        srcDir,
		PagesDir:      pagesDir,
		OutDir:        outDir,
		PublicDir:     publicDir,
		Router:        router.NewRouter(pagesDir),
		Bundler:       bundler,
		Compiler:      compiler.NewComponentCompiler(baseDir),
		PluginManager: pluginMgr,
	}
}

func (b *SSGBuilder) Build() error {
	baseDir := b.SrcDir
	if err := b.PluginManager.Load(baseDir, b.OutDir); err != nil {
		return fmt.Errorf("load plugins: %w", err)
	}

	buildCtx := &plugins.BuildContext{
		Config:    b.Config,
		RootDir:   baseDir,
		OutDir:    b.OutDir,
		PagesDir:  b.PagesDir,
		PublicDir: b.PublicDir,
	}

	if err := b.PluginManager.BuildStart(buildCtx); err != nil {
		return fmt.Errorf("plugin BuildStart: %w", err)
	}

	if err := b.Router.Discover(); err != nil {
		return fmt.Errorf("route discovery: %w", err)
	}
	b.Router.Sort()

	if err := os.RemoveAll(b.OutDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clean output: %w", err)
	}

	if err := os.MkdirAll(b.OutDir, 0755); err != nil {
		return fmt.Errorf("create output: %w", err)
	}

	for _, route := range b.Router.Routes {
		if route.Type == router.RouteStatic {
			if err := b.buildStaticRoute(route); err != nil {
				return fmt.Errorf("build %s: %w", route.Pattern, err)
			}
		}
	}

	if err := b.copyPublicAssets(); err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}

	if err := b.PluginManager.BuildEnd(buildCtx); err != nil {
		return fmt.Errorf("plugin BuildEnd: %w", err)
	}

	return nil
}

func (b *SSGBuilder) buildStaticRoute(route *router.Route) error {
	content, err := os.ReadFile(route.FilePath)
	if err != nil {
		return err
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		return err
	}

	resolver := b.Compiler.Resolver
	resolver.SetCurrentFile(route.FilePath)

	imports := make([]compiler.Import, len(comp.Imports))
	for i, imp := range comp.Imports {
		imports[i] = compiler.Import{
			Path:        imp.Path,
			Alias:       imp.Alias,
			IsComponent: imp.IsComponent,
		}
	}
	resolver.ParseImports(imports)

	ctx := executor.NewContext()
	if comp.Frontmatter != "" {
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			return err
		}
	}

	b.Compiler.CollectedStyles = nil
	processedTemplate := b.Compiler.ProcessComponentTags(comp.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		return err
	}

	allStyles := append(comp.Styles, b.Compiler.CollectedStyles...)
	compWithStyles := &parser.Component{
		Frontmatter: comp.Frontmatter,
		Template:    comp.Template,
		Scripts:     comp.Scripts,
		Styles:      allStyles,
		Imports:     comp.Imports,
	}

	cssPath, err := b.Bundler.BundleStyles(compWithStyles, route.FilePath)
	if err != nil {
		return err
	}

	jsPath, err := b.Bundler.BundleScripts(comp, route.FilePath)
	if err != nil {
		return err
	}

	scopeID := ""
	for _, style := range allStyles {
		if style.Scoped {
			scopeID = b.Bundler.GenerateScopeID(route.FilePath)
			break
		}
	}

	rendered = b.Bundler.InjectAssets(rendered, cssPath, jsPath, scopeID)

	outPath := b.getOutputPath(route.Pattern)
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(outPath, []byte(rendered), 0644); err != nil {
		return err
	}

	fmt.Printf("  ✓ %s → %s\n", route.Pattern, outPath)
	return nil
}

func (b *SSGBuilder) getOutputPath(pattern string) string {
	if pattern == "/" {
		return filepath.Join(b.OutDir, "index.html")
	}

	pattern = strings.TrimPrefix(pattern, "/")
	return filepath.Join(b.OutDir, pattern, "index.html")
}

func (b *SSGBuilder) copyPublicAssets() error {
	if _, err := os.Stat(b.PublicDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(b.PublicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(b.PublicDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(b.OutDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}
