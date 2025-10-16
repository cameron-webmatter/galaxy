package server

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/galaxy/galaxy/internal/assets"
	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/endpoints"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/lifecycle"
	"github.com/galaxy/galaxy/pkg/middleware"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/ssr"
	"github.com/galaxy/galaxy/pkg/template"
)

type DevServer struct {
	Router           *router.Router
	RootDir          string
	PagesDir         string
	PublicDir        string
	Port             int
	Bundler          *assets.Bundler
	Compiler         *compiler.ComponentCompiler
	EndpointCompiler *endpoints.EndpointCompiler
	Verbose          bool
	Lifecycle        *lifecycle.Lifecycle
}

func NewDevServer(rootDir, pagesDir, publicDir string, port int, verbose bool) *DevServer {
	srcDir := filepath.Dir(pagesDir)
	return &DevServer{
		Router:           router.NewRouter(pagesDir),
		RootDir:          rootDir,
		PagesDir:         pagesDir,
		PublicDir:        publicDir,
		Port:             port,
		Bundler:          assets.NewBundler(".galaxy"),
		Compiler:         compiler.NewComponentCompiler(srcDir),
		EndpointCompiler: endpoints.NewCompiler(rootDir, ".galaxy/endpoints"),
		Verbose:          verbose,
	}
}

func (s *DevServer) Start() error {
	if err := s.Router.Discover(); err != nil {
		return err
	}
	s.Router.Sort()

	if s.Lifecycle != nil {
		if err := s.Lifecycle.ExecuteStartup(); err != nil {
			return fmt.Errorf("lifecycle startup: %w", err)
		}
	}

	http.HandleFunc("/", s.logRequest(s.handleRequest))

	addr := fmt.Sprintf(":%d", s.Port)
	fmt.Printf("🚀 Dev server running at http://localhost%s\n", addr)
	fmt.Printf("📁 Pages: %s\n", s.PagesDir)
	fmt.Printf("📦 Public: %s\n\n", s.PublicDir)

	s.printRoutes()

	return http.ListenAndServe(addr, nil)
}

func (s *DevServer) ReloadRoutes() error {
	if err := s.Router.Reload(); err != nil {
		return err
	}

	fmt.Println("\n🔄 Routes reloaded:")
	s.printRoutes()

	return nil
}

func (s *DevServer) printRoutes() {
	for _, route := range s.Router.Routes {
		fmt.Printf("  %s\n", route.Pattern)
	}
	fmt.Println()
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *DevServer) logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		next(rw, r)

		if s.Verbose {
			duration := time.Since(start)
			statusColor := getStatusColor(rw.statusCode)
			methodColor := "\033[36m"
			reset := "\033[0m"

			fmt.Printf("%s%s%s %s - %s%d%s (%dms)\n",
				methodColor, r.Method, reset,
				r.URL.Path,
				statusColor, rw.statusCode, reset,
				duration.Milliseconds())
		}
	}
}

func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return "\033[31m"
	case status >= 400:
		return "\033[33m"
	case status >= 300:
		return "\033[36m"
	case status >= 200:
		return "\033[32m"
	default:
		return "\033[0m"
	}
}

func (s *DevServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/wasm_exec.js" {
		s.serveWasmExec(w, r)
		return
	}

	if filepath.Ext(r.URL.Path) != "" {
		s.serveStatic(w, r)
		return
	}

	route, params := s.Router.Match(r.URL.Path)
	if route == nil {
		http.NotFound(w, r)
		return
	}

	mwCtx := middleware.NewContext(w, r)
	mwCtx.Params = params

	if route.IsEndpoint {
		s.handleEndpoint(route, mwCtx, params)
		return
	}

	s.handlePage(route, mwCtx, params)
}

func (s *DevServer) handleEndpoint(route *router.Route, mwCtx *middleware.Context, params map[string]string) {
	endpoint, err := s.EndpointCompiler.Load(route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Load endpoint: %v", err), http.StatusInternalServerError)
		return
	}

	if err := endpoints.HandleEndpoint(endpoint, mwCtx.Response, mwCtx.Request, params, mwCtx.Locals); err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
	}
}

func (s *DevServer) handlePage(route *router.Route, mwCtx *middleware.Context, params map[string]string) {
	content, err := os.ReadFile(route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, err.Error(), http.StatusInternalServerError)
		return
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Parse error: %v", err), http.StatusInternalServerError)
		return
	}

	resolver := s.Compiler.Resolver
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

	reqCtx := ssr.NewRequestContext(mwCtx.Request, params)
	ctx.SetRequest(reqCtx)
	ctx.SetLocals(mwCtx.Locals)

	for k, v := range params {
		ctx.Set(k, v)
	}

	if comp.Frontmatter != "" {
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			http.Error(mwCtx.Response, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if ctx.ShouldRedirect {
		http.Redirect(mwCtx.Response, mwCtx.Request, ctx.RedirectURL, ctx.RedirectStatus)
		return
	}

	s.Compiler.CollectedStyles = nil
	processedTemplate := s.Compiler.ProcessComponentTags(comp.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	allStyles := append(comp.Styles, s.Compiler.CollectedStyles...)
	compWithStyles := &parser.Component{
		Frontmatter: comp.Frontmatter,
		Template:    comp.Template,
		Scripts:     comp.Scripts,
		Styles:      allStyles,
		Imports:     comp.Imports,
	}

	cssPath, err := s.Bundler.BundleStyles(compWithStyles, route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Style bundle error: %v", err), http.StatusInternalServerError)
		return
	}

	jsPath, err := s.Bundler.BundleScripts(comp, route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("Script bundle error: %v", err), http.StatusInternalServerError)
		return
	}

	wasmAssets, err := s.Bundler.BundleWasmScripts(comp, route.FilePath)
	if err != nil {
		http.Error(mwCtx.Response, fmt.Sprintf("WASM bundle error: %v", err), http.StatusInternalServerError)
		return
	}

	scopeID := ""
	for _, style := range allStyles {
		if style.Scoped {
			scopeID = s.Bundler.GenerateScopeID(route.FilePath)
			break
		}
	}

	rendered = s.Bundler.InjectAssetsWithWasm(rendered, cssPath, jsPath, scopeID, wasmAssets)

	mwCtx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	mwCtx.Response.Write([]byte(rendered))
}

func (s *DevServer) serveStatic(w http.ResponseWriter, r *http.Request) {
	galaxyPath := filepath.Join(".galaxy", r.URL.Path)
	if _, err := os.Stat(galaxyPath); err == nil {
		http.ServeFile(w, r, galaxyPath)
		return
	}

	publicPath := filepath.Join(s.PublicDir, r.URL.Path)
	http.ServeFile(w, r, publicPath)
}

func (s *DevServer) serveWasmExec(w http.ResponseWriter, r *http.Request) {
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		cmd := exec.Command("go", "env", "GOROOT")
		output, _ := cmd.Output()
		goRoot = strings.TrimSpace(string(output))
	}

	wasmExecPath := filepath.Join(goRoot, "misc", "wasm", "wasm_exec.js")
	if _, err := os.Stat(wasmExecPath); os.IsNotExist(err) {
		wasmExecPath = filepath.Join(goRoot, "lib", "wasm", "wasm_exec.js")
	}

	http.ServeFile(w, r, wasmExecPath)
}
