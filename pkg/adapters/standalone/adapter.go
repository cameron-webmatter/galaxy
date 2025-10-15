package standalone

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/galaxy/galaxy/pkg/adapters"
)

type StandaloneAdapter struct{}

func New() *StandaloneAdapter {
	return &StandaloneAdapter{}
}

func (a *StandaloneAdapter) Name() string {
	return "standalone"
}

func (a *StandaloneAdapter) Build(cfg *adapters.BuildConfig) error {
	if err := a.generateMain(cfg); err != nil {
		return fmt.Errorf("generate main: %w", err)
	}

	if err := a.generateGoMod(cfg); err != nil {
		return fmt.Errorf("generate go.mod: %w", err)
	}

	if err := a.compile(cfg); err != nil {
		return fmt.Errorf("compile: %w", err)
	}

	return nil
}

func (a *StandaloneAdapter) generateMain(cfg *adapters.BuildConfig) error {
	mainPath := filepath.Join(cfg.ServerDir, "main.go")

	if err := a.copyProjectFiles(cfg); err != nil {
		return fmt.Errorf("copy project files: %w", err)
	}

	endpoints := a.buildEndpointData(cfg)
	imports := a.buildEndpointImports(cfg)
	hasMiddleware := a.checkMiddleware(cfg)
	hasSequence := a.checkSequence(cfg)

	tmpl := template.Must(template.New("main").Parse(mainTemplate))

	f, err := os.Create(mainPath)
	if err != nil {
		return err
	}
	defer f.Close()

	routes := []map[string]interface{}{}
	for _, r := range cfg.Routes {
		relPath, _ := filepath.Rel(cfg.PagesDir, r.FilePath)
		routes = append(routes, map[string]interface{}{
			"Pattern":    r.Pattern,
			"RelPath":    "/" + relPath,
			"IsEndpoint": r.IsEndpoint,
		})
	}

	data := map[string]interface{}{
		"Port":            cfg.Config.Server.Port,
		"Host":            cfg.Config.Server.Host,
		"PublicDir":       filepath.Join(cfg.OutDir, "public"),
		"StaticDir":       cfg.OutDir,
		"PagesDir":        cfg.PagesDir,
		"Routes":          routes,
		"Endpoints":       endpoints,
		"EndpointImports": imports,
		"HasMiddleware":   hasMiddleware,
		"HasSequence":     hasSequence,
	}

	return tmpl.Execute(f, data)
}

func (a *StandaloneAdapter) buildEndpointData(cfg *adapters.BuildConfig) []map[string]interface{} {
	endpoints := []map[string]interface{}{}

	for _, route := range cfg.Routes {
		if !route.IsEndpoint {
			continue
		}

		pkgName := a.getPackageName(route.FilePath)
		methods := a.detectMethods(route.FilePath, pkgName)

		if len(methods) > 0 {
			endpoints = append(endpoints, map[string]interface{}{
				"Pattern": route.Pattern,
				"Methods": methods,
				"Package": pkgName,
			})
		}
	}

	return endpoints
}

func (a *StandaloneAdapter) detectMethods(filePath, pkgName string) []map[string]string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	src := string(content)
	methods := []map[string]string{}

	httpMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "ALL"}
	for _, method := range httpMethods {
		pattern := fmt.Sprintf("func %s(", method)
		if contains(src, pattern) {
			methods = append(methods, map[string]string{
				"Method":  method,
				"Package": pkgName,
			})
		}
	}

	return methods
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (a *StandaloneAdapter) buildEndpointImports(cfg *adapters.BuildConfig) []map[string]string {
	imports := []map[string]string{}
	seen := make(map[string]bool)

	for _, route := range cfg.Routes {
		if !route.IsEndpoint {
			continue
		}

		pkgName := a.getPackageName(route.FilePath)
		if seen[pkgName] {
			continue
		}
		seen[pkgName] = true

		relPath, _ := filepath.Rel(cfg.PagesDir, filepath.Dir(route.FilePath))
		importPath := filepath.Join("galaxy-server/pages", relPath)

		imports = append(imports, map[string]string{
			"Alias": pkgName,
			"Path":  importPath,
		})
	}

	return imports
}

func (a *StandaloneAdapter) getPackageName(filePath string) string {
	dir := filepath.Dir(filePath)
	base := filepath.Base(dir)
	if base == "api" || base == "." {
		return "api"
	}
	return base
}

func (a *StandaloneAdapter) checkMiddleware(cfg *adapters.BuildConfig) bool {
	projectDir := filepath.Dir(cfg.PagesDir)
	middlewarePath := filepath.Join(projectDir, "src", "middleware.go")
	_, err := os.Stat(middlewarePath)
	return err == nil
}

func (a *StandaloneAdapter) checkSequence(cfg *adapters.BuildConfig) bool {
	projectDir := filepath.Dir(cfg.PagesDir)
	middlewarePath := filepath.Join(projectDir, "src", "middleware.go")
	content, err := os.ReadFile(middlewarePath)
	if err != nil {
		return false
	}
	return contains(string(content), "func Sequence()")
}

func (a *StandaloneAdapter) copyProjectFiles(cfg *adapters.BuildConfig) error {
	pagesOutDir := filepath.Join(cfg.ServerDir, "pages")
	if err := os.MkdirAll(pagesOutDir, 0755); err != nil {
		return err
	}

	projectDir := filepath.Dir(cfg.PagesDir)

	srcDir := filepath.Join(projectDir, "src")
	if _, err := os.Stat(srcDir); err == nil {
		srcOut := filepath.Join(cfg.ServerDir, "src")
		if err := os.MkdirAll(srcOut, 0755); err != nil {
			return err
		}

		filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(srcDir, path)
			destPath := filepath.Join(srcOut, relPath)
			os.MkdirAll(filepath.Dir(destPath), 0755)
			data, _ := os.ReadFile(path)
			return os.WriteFile(destPath, data, 0644)
		})
	}

	componentsDir := projectDir
	componentsPath := filepath.Join(componentsDir, "components")

	if _, err := os.Stat(componentsPath); err == nil {
		componentsOut := filepath.Join(cfg.ServerDir, "components")
		if err := os.MkdirAll(componentsOut, 0755); err != nil {
			return err
		}

		filepath.Walk(componentsPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(componentsPath, path)
			destPath := filepath.Join(componentsOut, relPath)
			os.MkdirAll(filepath.Dir(destPath), 0755)
			data, _ := os.ReadFile(path)
			return os.WriteFile(destPath, data, 0644)
		})
	}

	return filepath.Walk(cfg.PagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(cfg.PagesDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(pagesOutDir, relPath)

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0644)
	})
}

func (a *StandaloneAdapter) generateGoMod(cfg *adapters.BuildConfig) error {
	modPath := filepath.Join(cfg.ServerDir, "go.mod")

	galaxyPath, err := a.getGalaxyModulePath()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(`module galaxy-server

go 1.21

replace github.com/galaxy/galaxy => %s

require github.com/galaxy/galaxy v0.0.0
`, galaxyPath)

	return os.WriteFile(modPath, []byte(content), 0644)
}

func (a *StandaloneAdapter) getGalaxyModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/galaxy/galaxy")
	out, err := cmd.Output()
	if err != nil {
		wd, _ := os.Getwd()
		for wd != "/" {
			modPath := filepath.Join(wd, "go.mod")
			if _, err := os.Stat(modPath); err == nil {
				return wd, nil
			}
			wd = filepath.Dir(wd)
		}
		return "", fmt.Errorf("cannot find galaxy module")
	}
	return string(out), nil
}

func (a *StandaloneAdapter) compile(cfg *adapters.BuildConfig) error {
	binaryPath := filepath.Join(cfg.ServerDir, "server")

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = cfg.ServerDir
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = cfg.ServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

const mainTemplate = `package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/endpoints"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/middleware"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/ssr"
	"github.com/galaxy/galaxy/pkg/template"

	{{range .EndpointImports}}
	{{.Alias}} "{{.Path}}"
	{{end}}
	{{if .HasMiddleware}}
	usermw "galaxy-server/src"
	{{end}}
)

var (
	rt       *router.Router
	comp     *compiler.ComponentCompiler
	baseDir  string
	pagesDir = "pages"
	routeMap = map[string]*routeInfo{
		{{range .Routes}}
		"{{.Pattern}}": {Pattern: "{{.Pattern}}", FilePath: "pages{{.RelPath}}", IsEndpoint: {{.IsEndpoint}}},
		{{end}}
	}
	endpointHandlers = map[string]map[string]endpoints.HandlerFunc{
		{{range .Endpoints}}
		"{{.Pattern}}": {
			{{range .Methods}}
			"{{.Method}}": {{.Package}}.{{.Method}},
			{{end}}
		},
		{{end}}
	}
)

type routeInfo struct {
	Pattern    string
	FilePath   string
	IsEndpoint bool
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir = filepath.Dir(exePath)
	comp = compiler.NewComponentCompiler(baseDir)

	http.HandleFunc("/", handleRequest)

	addr := "{{.Host}}:{{.Port}}"
	log.Printf("🚀 Server running at http://%s\n", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if filepath.Ext(r.URL.Path) != "" {
		http.ServeFile(w, r, filepath.Join("{{.PublicDir}}", r.URL.Path))
		return
	}

	staticPath := filepath.Join("{{.StaticDir}}", r.URL.Path)
	if r.URL.Path == "/" {
		staticPath = filepath.Join("{{.StaticDir}}", "index.html")
	} else {
		staticPath = filepath.Join("{{.StaticDir}}", r.URL.Path, "index.html")
	}
	
	if _, err := os.Stat(staticPath); err == nil {
		http.ServeFile(w, r, staticPath)
		return
	}

	route, ok := routeMap[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	mwCtx := middleware.NewContext(w, r)
	mwCtx.Params = make(map[string]string)

	{{if .HasMiddleware}}
	{{if .HasSequence}}
	chain := middleware.NewChain()
	for _, mw := range usermw.Sequence() {
		chain.Use(mw)
	}
	if err := chain.Execute(mwCtx, func(ctx *middleware.Context) error {
		if route.IsEndpoint {
			handleEndpoint(route, mwCtx)
		} else {
			handlePage(route, mwCtx)
		}
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	{{else}}
	if err := usermw.OnRequest(mwCtx, func() error {
		if route.IsEndpoint {
			handleEndpoint(route, mwCtx)
		} else {
			handlePage(route, mwCtx)
		}
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	{{end}}
	{{else}}
	if route.IsEndpoint {
		handleEndpoint(route, mwCtx)
		return
	}
	handlePage(route, mwCtx)
	{{end}}
}

func handleEndpoint(route *routeInfo, mwCtx *middleware.Context) {
	ep, ok := endpointHandlers[route.Pattern]
	if !ok {
		http.Error(mwCtx.Response, "Endpoint not found", http.StatusNotFound)
		return
	}

	method := mwCtx.Request.Method
	handler, ok := ep[method]
	if !ok {
		handler, ok = ep["ALL"]
		if !ok {
			http.Error(mwCtx.Response, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	ctx := endpoints.NewContext(mwCtx.Response, mwCtx.Request, mwCtx.Params, mwCtx.Locals)
	if err := handler(ctx); err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
	}
}

func handlePage(route *routeInfo, mwCtx *middleware.Context) {
	filePath := filepath.Join(baseDir, route.FilePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
		return
	}

	parsed, err := parser.Parse(string(content))
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Parse error: %v", err), http.StatusInternalServerError)
		return
	}

	resolver := comp.Resolver
	resolver.SetCurrentFile(filePath)

	imports := make([]compiler.Import, len(parsed.Imports))
	for i, imp := range parsed.Imports {
		imports[i] = compiler.Import{
			Path:        imp.Path,
			Alias:       imp.Alias,
			IsComponent: imp.IsComponent,
		}
	}
	resolver.ParseImports(imports)

	ctx := executor.NewContext()

	reqCtx := ssr.NewRequestContext(mwCtx.Request, mwCtx.Params)
	ctx.SetRequest(reqCtx)
	ctx.SetLocals(mwCtx.Locals)

	for k, v := range mwCtx.Params {
		ctx.Set(k, v)
	}

	if parsed.Frontmatter != "" {
		if err := ctx.Execute(parsed.Frontmatter); err != nil {
			http.Error(mwCtx.Response, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	comp.CollectedStyles = nil
	processedTemplate := comp.ProcessComponentTags(parsed.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	allStyles := append(parsed.Styles, comp.CollectedStyles...)
	if len(allStyles) > 0 {
		var styleContent string
		for _, style := range allStyles {
			styleContent += style.Content + "\n"
		}
		styleTag := "<style>" + styleContent + "</style>"
		rendered = strings.Replace(rendered, "</head>", styleTag+"\n</head>", 1)
	}

	mwCtx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	mwCtx.Response.Write([]byte(rendered))
}
`
