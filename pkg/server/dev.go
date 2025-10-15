package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/galaxy/galaxy/internal/assets"
	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/parser"
	"github.com/galaxy/galaxy/pkg/router"
	"github.com/galaxy/galaxy/pkg/ssr"
	"github.com/galaxy/galaxy/pkg/template"
)

type DevServer struct {
	Router    *router.Router
	PagesDir  string
	PublicDir string
	Port      int
	Bundler   *assets.Bundler
	Compiler  *compiler.ComponentCompiler
	Verbose   bool
}

func NewDevServer(pagesDir, publicDir string, port int, verbose bool) *DevServer {
	baseDir := filepath.Dir(pagesDir)
	return &DevServer{
		Router:    router.NewRouter(pagesDir),
		PagesDir:  pagesDir,
		PublicDir: publicDir,
		Port:      port,
		Bundler:   assets.NewBundler(".tmp"),
		Compiler:  compiler.NewComponentCompiler(baseDir),
		Verbose:   verbose,
	}
}

func (s *DevServer) Start() error {
	if err := s.Router.Discover(); err != nil {
		return err
	}
	s.Router.Sort()

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
	if filepath.Ext(r.URL.Path) != "" {
		s.serveStatic(w, r)
		return
	}

	route, params := s.Router.Match(r.URL.Path)
	if route == nil {
		http.NotFound(w, r)
		return
	}

	content, err := os.ReadFile(route.FilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusInternalServerError)
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

	reqCtx := ssr.NewRequestContext(r, params)
	ctx.SetRequest(reqCtx)

	for k, v := range params {
		ctx.Set(k, v)
	}

	if comp.Frontmatter != "" {
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	processedTemplate := s.Compiler.ProcessComponentTags(comp.Template, ctx)

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(processedTemplate, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(rendered))
}

func (s *DevServer) serveStatic(w http.ResponseWriter, r *http.Request) {
	publicPath := filepath.Join(s.PublicDir, r.URL.Path)

	if _, err := os.Stat(publicPath); err == nil {
		http.ServeFile(w, r, publicPath)
		return
	}

	http.NotFound(w, r)
}
