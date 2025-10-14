package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gastro/gastro/internal/assets"
	"github.com/gastro/gastro/pkg/executor"
	"github.com/gastro/gastro/pkg/parser"
	"github.com/gastro/gastro/pkg/router"
	"github.com/gastro/gastro/pkg/template"
)

type DevServer struct {
	Router    *router.Router
	PagesDir  string
	PublicDir string
	Port      int
	Bundler   *assets.Bundler
}

func NewDevServer(pagesDir, publicDir string, port int) *DevServer {
	return &DevServer{
		Router:    router.NewRouter(pagesDir),
		PagesDir:  pagesDir,
		PublicDir: publicDir,
		Port:      port,
		Bundler:   assets.NewBundler(".tmp"),
	}
}

func (s *DevServer) Start() error {
	if err := s.Router.Discover(); err != nil {
		return err
	}
	s.Router.Sort()

	http.HandleFunc("/", s.handleRequest)

	addr := fmt.Sprintf(":%d", s.Port)
	fmt.Printf("🚀 Dev server running at http://localhost%s\n", addr)
	fmt.Printf("📁 Pages: %s\n", s.PagesDir)
	fmt.Printf("📦 Public: %s\n\n", s.PublicDir)

	for _, route := range s.Router.Routes {
		fmt.Printf("  %s\n", route.Pattern)
	}

	return http.ListenAndServe(addr, nil)
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

	ctx := executor.NewContext()
	for k, v := range params {
		ctx.Set(k, v)
	}

	if comp.Frontmatter != "" {
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			http.Error(w, fmt.Sprintf("Execution error: %v", err), http.StatusInternalServerError)
			return
		}
	}

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(comp.Template, nil)
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
