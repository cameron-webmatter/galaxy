package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	
	"generated-hybrid-ssg/runtime"
)

func main() {
	fmt.Println("Pre-rendering pages...")
	
	renderPage("/", "/Users/cameron/dev/galaxy/examples/hybrid/dist/index.html", HandleIndex)
	
	fmt.Println("✓ Done")
}

func renderPage(pattern, outPath string, handler func(http.ResponseWriter, *http.Request, map[string]string)) {
	w := &responseWriter{body: make([]byte, 0)}
	r := &http.Request{}
	params := make(map[string]string)
	
	handler(w, r, params)
	
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		panic(err)
	}
	
	if err := os.WriteFile(outPath, w.body, 0644); err != nil {
		panic(err)
	}
	
	fmt.Printf("  ✓ %s -> %s\n", pattern, outPath)
}

type responseWriter struct {
	body []byte
}

func (w *responseWriter) Header() http.Header { return http.Header{} }
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}
func (w *responseWriter) WriteHeader(statusCode int) {}

func HandleIndex(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var title = "Static Home"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/index.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleIndex)
	w.Write([]byte(html))
}

const templateHandleIndex = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>Static Page (Prerendered)</h1>
    <p>This page is built at compile time.</p>
    <nav>
        <a href="/dynamic">Dynamic Page</a>
    </nav>
</body>
</html>`

