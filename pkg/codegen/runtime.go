package codegen

import "fmt"

func (g *MainGenerator) GenerateRuntime() string {
	return fmt.Sprintf(`package runtime

import (
	"encoding/json"
	"os"
	"strings"
	
	"github.com/galaxy/galaxy/pkg/compiler"
	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/template"
	"github.com/galaxy/galaxy/pkg/wasm"
)

var comp *compiler.ComponentCompiler
var wasmManifest *wasm.WasmManifest

func init() {
	comp = compiler.NewComponentCompiler(".")
	loadWasmManifest()
}

func loadWasmManifest() {
	data, err := os.ReadFile(%q)
	if err != nil {
		return
	}
	wasmManifest = &wasm.WasmManifest{}
	json.Unmarshal(data, wasmManifest)
}

type RenderContext struct {
	*executor.Context
	RoutePath string
}

func NewRenderContext() *RenderContext {
	return &RenderContext{
		Context: executor.NewContext(),
	}
}

func RenderTemplate(ctx *RenderContext, templateHTML string) string {
	processed := comp.ProcessComponentTags(templateHTML, ctx.Context)
	
	engine := template.NewEngine(ctx.Context)
	rendered, _ := engine.Render(processed, nil)
	
	rendered = injectWasmScripts(rendered, ctx.RoutePath)
	
	return rendered
}

func injectWasmScripts(html, routePath string) string {
	if wasmManifest == nil {
		return html
	}
	
	assets, ok := wasmManifest.Assets[routePath]
	if !ok {
		return html
	}
	
	var scripts []string
	
	for _, mod := range assets.WasmModules {
		scripts = append(scripts, 
			"<script src=\"/wasm_exec.js\"></script>",
			"<script src=\"" + mod.LoaderPath + "\"></script>")
	}
	
	for _, js := range assets.JSScripts {
		scripts = append(scripts, "<script src=\"" + js + "\"></script>")
	}
	
	if len(scripts) > 0 {
		injection := strings.Join(scripts, "\n")
		html = strings.Replace(html, "</body>", injection + "\n</body>", 1)
	}
	
	return html
}
`, g.ManifestPath)
}
