package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	
	"generated-ssg/runtime"
)

func main() {
	fmt.Println("Pre-rendering pages...")
	
	renderPage("/", "/Users/cameron/dev/galaxy/examples/basic/dist/index.html", HandleIndex)
	renderPage("/about", "/Users/cameron/dev/galaxy/examples/basic/dist/about/index.html", HandleAbout)
	renderPage("/components-test", "/Users/cameron/dev/galaxy/examples/basic/dist/components-test/index.html", HandleComponentsTest)
	renderPage("/ssr-demo", "/Users/cameron/dev/galaxy/examples/basic/dist/ssr-demo/index.html", HandleSsrDemo)
	renderPage("/wasm-simple", "/Users/cameron/dev/galaxy/examples/basic/dist/wasm-simple/index.html", HandleWasmSimple)
	renderPage("/wasm-test", "/Users/cameron/dev/galaxy/examples/basic/dist/wasm-test/index.html", HandleWasmTest)
	
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
	
	
	var title = "Welcome to Galaxy"
var description = "A Go-powered web framework"
var features = []string{"Fast", "Type-safe", "Go runtime"}
var count = 3
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/index.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	ctx.Set("description", description)
	ctx.Set("features", features)
	ctx.Set("count", count)
	
	html := runtime.RenderTemplate(ctx, templateHandleIndex)
	w.Write([]byte(html))
}

const templateHandleIndex = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
</head>
<body>
    <header>
        <h1>{title}</h1>
        <p>{description}</p>
    </header>

    <main>
        <section>
            <h2>Features</h2>
            <ul>
                <li galaxy:for={feature in features}>{feature}</li>
            </ul>
        </section>

        <div galaxy:if={count}>
            <p>We have {count} amazing features!</p>
        </div>
    </main>

    <footer>
        <p>Built with Galaxy</p>
    </footer>
</body>
</html>`


func HandleAbout(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var pageTitle = "About Galaxy"
var version = "0.1.0"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/about.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("pageTitle", pageTitle)
	ctx.Set("version", version)
	
	html := runtime.RenderTemplate(ctx, templateHandleAbout)
	w.Write([]byte(html))
}

const templateHandleAbout = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{pageTitle}</title>
</head>
<body>
    <h1>{pageTitle}</h1>
    <p>Version: {version}</p>
    <a href="/">Back to Home</a>
</body>
</html>`


func HandleComponentsTest(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var cardTitle = "My Card"
var cardDesc = "This is a card component"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/components-test.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("cardTitle", cardTitle)
	ctx.Set("cardDesc", cardDesc)
	
	html := runtime.RenderTemplate(ctx, templateHandleComponentsTest)
	w.Write([]byte(html))
}

const templateHandleComponentsTest = `<!DOCTYPE html>
<html>
<head>
    <title>Component Test</title>
</head>
<body>
    <h1>Component Composition Test</h1>
    
    <Card title={cardTitle} description={cardDesc}>
        <p>This is slot content!</p>
    </Card>
</body>
</html>`


func HandleSsrDemo(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var title = "SSR Demo"
var serverTime = "Server rendered at request time"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/ssr-demo.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	ctx.Set("serverTime", serverTime)
	
	html := runtime.RenderTemplate(ctx, templateHandleSsrDemo)
	w.Write([]byte(html))
}

const templateHandleSsrDemo = `<!DOCTYPE html>
<html>
<head>
    <title>{title}</title>
</head>
<body>
    <h1>{title}</h1>
    
    <div galaxy:if={Request}>
        <h2>Request Information</h2>
        <p>Path: {Request.Path()}</p>
        <p>Method: {Request.Method()}</p>
        <p>URL: {Request.URL()}</p>
    </div>

    <p>{serverTime}</p>
</body>
</html>`


func HandleWasmSimple(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var title = "Simple WASM Demo"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/wasm-simple.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleWasmSimple)
	w.Write([]byte(html))
}

const templateHandleWasmSimple = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{title}</title>
</head>
<body>
    <h1>{title}</h1>
    <button id="btn">Click Me</button>
    <p id="output">Count: 0</p>
</body>
</html>`


func HandleWasmTest(w http.ResponseWriter, r *http.Request, params map[string]string) {
	
	
	var title = "WASM Counter Test"
	
	ctx := runtime.NewRenderContext()
	ctx.RoutePath = "pages/wasm-test.gxc"
	for k, v := range params {
		ctx.Set(k, v)
	}
	
		ctx.Set("title", title)
	
	html := runtime.RenderTemplate(ctx, templateHandleWasmTest)
	w.Write([]byte(html))
}

const templateHandleWasmTest = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{title}</title>
</head>
<body>
    <header>
        <h1>{title}</h1>
        <p>Click the button to test Go WASM</p>
    </header>

    <main>
        <div id="counter">
            <button id="increment">Increment</button>
            <span id="count">0</span>
        </div>
    </main>
</body>
</html>`

