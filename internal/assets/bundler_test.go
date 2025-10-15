package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/galaxy/galaxy/pkg/parser"
)

func TestNewBundler(t *testing.T) {
	bundler := NewBundler("/out")

	if bundler.OutDir != "/out" {
		t.Errorf("Expected OutDir /out, got %s", bundler.OutDir)
	}
}

func TestBundleStylesEmpty(t *testing.T) {
	bundler := NewBundler(t.TempDir())
	comp := &parser.Component{
		Styles: []parser.Style{},
	}

	path, err := bundler.BundleStyles(comp, "/pages/index.gxc")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if path != "" {
		t.Errorf("Expected empty path for no styles, got %s", path)
	}
}

func TestBundleStyles(t *testing.T) {
	tmpDir := t.TempDir()
	bundler := NewBundler(tmpDir)

	comp := &parser.Component{
		Styles: []parser.Style{
			{Content: ".header { color: blue; }", Scoped: false},
			{Content: ".footer { color: red; }", Scoped: false},
		},
	}

	path, err := bundler.BundleStyles(comp, "/pages/index.gxc")
	if err != nil {
		t.Fatalf("BundleStyles failed: %v", err)
	}

	if !strings.HasPrefix(path, "/_assets/styles-") {
		t.Errorf("Expected path to start with /_assets/styles-, got %s", path)
	}

	if !strings.HasSuffix(path, ".css") {
		t.Errorf("Expected path to end with .css, got %s", path)
	}

	filePath := filepath.Join(tmpDir, strings.TrimPrefix(path, "/"))
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read bundled file: %v", err)
	}

	if !strings.Contains(string(content), ".header") {
		t.Error("Expected bundled CSS to contain .header")
	}

	if !strings.Contains(string(content), ".footer") {
		t.Error("Expected bundled CSS to contain .footer")
	}
}

func TestBundleStylesScoped(t *testing.T) {
	tmpDir := t.TempDir()
	bundler := NewBundler(tmpDir)

	comp := &parser.Component{
		Styles: []parser.Style{
			{Content: ".container { padding: 10px; }", Scoped: true},
		},
	}

	path, err := bundler.BundleStyles(comp, "/pages/test.gxc")
	if err != nil {
		t.Fatalf("BundleStyles failed: %v", err)
	}

	filePath := filepath.Join(tmpDir, strings.TrimPrefix(path, "/"))
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read bundled file: %v", err)
	}

	scopeID := bundler.GenerateScopeID("/pages/test.gxc")
	expectedAttr := "[data-gxc-" + scopeID + "]"

	if !strings.Contains(string(content), expectedAttr) {
		t.Errorf("Expected scoped CSS to contain %s, got %s", expectedAttr, string(content))
	}
}

func TestBundleScriptsEmpty(t *testing.T) {
	bundler := NewBundler(t.TempDir())
	comp := &parser.Component{
		Scripts: []parser.Script{},
	}

	path, err := bundler.BundleScripts(comp, "/pages/index.gxc")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if path != "" {
		t.Errorf("Expected empty path for no scripts, got %s", path)
	}
}

func TestBundleScripts(t *testing.T) {
	tmpDir := t.TempDir()
	bundler := NewBundler(tmpDir)

	comp := &parser.Component{
		Scripts: []parser.Script{
			{Content: "console.log('hello');", IsModule: false},
			{Content: "console.log('world');", IsModule: true},
		},
	}

	path, err := bundler.BundleScripts(comp, "/pages/index.gxc")
	if err != nil {
		t.Fatalf("BundleScripts failed: %v", err)
	}

	if !strings.HasPrefix(path, "/_assets/script-") {
		t.Errorf("Expected path to start with /_assets/script-, got %s", path)
	}

	if !strings.HasSuffix(path, ".js") {
		t.Errorf("Expected path to end with .js, got %s", path)
	}

	filePath := filepath.Join(tmpDir, strings.TrimPrefix(path, "/"))
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read bundled file: %v", err)
	}

	if !strings.Contains(string(content), "hello") {
		t.Error("Expected bundled JS to contain first script")
	}

	if !strings.Contains(string(content), "world") {
		t.Error("Expected bundled JS to contain second script")
	}
}

func TestScopeCSS(t *testing.T) {
	bundler := NewBundler("/out")

	css := `.header { color: red; }
.footer { color: blue; }`

	scoped := bundler.scopeCSS(css, "/pages/test.gxc")
	scopeID := bundler.GenerateScopeID("/pages/test.gxc")
	expectedAttr := "[data-gxc-" + scopeID + "]"

	if !strings.Contains(scoped, expectedAttr+" .header") {
		t.Errorf("Expected scoped selector with %s .header", expectedAttr)
	}

	if !strings.Contains(scoped, expectedAttr+" .footer") {
		t.Errorf("Expected scoped selector with %s .footer", expectedAttr)
	}
}

func TestScopeCSSWithComments(t *testing.T) {
	bundler := NewBundler("/out")

	css := `/* Comment */
.test { color: red; }`

	scoped := bundler.scopeCSS(css, "/page.gxc")

	if !strings.Contains(scoped, "/* Comment */") {
		t.Error("Expected comment to be preserved")
	}
}

func TestScopeCSSEmptyLines(t *testing.T) {
	bundler := NewBundler("/out")

	css := `.a { color: red; }

.b { color: blue; }`

	scoped := bundler.scopeCSS(css, "/page.gxc")
	scopeID := bundler.GenerateScopeID("/page.gxc")
	expectedAttr := "[data-gxc-" + scopeID + "]"

	if !strings.Contains(scoped, expectedAttr+" .a") {
		t.Error("Expected scoped .a selector")
	}

	if !strings.Contains(scoped, expectedAttr+" .b") {
		t.Error("Expected scoped .b selector")
	}
}

func TestGenerateScopeID(t *testing.T) {
	bundler := NewBundler("/out")

	id1 := bundler.GenerateScopeID("/pages/index.gxc")
	id2 := bundler.GenerateScopeID("/pages/index.gxc")

	if id1 != id2 {
		t.Error("Expected same path to generate same scope ID")
	}

	if len(id1) != 6 {
		t.Errorf("Expected scope ID length 6, got %d", len(id1))
	}

	id3 := bundler.GenerateScopeID("/pages/about.gxc")
	if id1 == id3 {
		t.Error("Expected different paths to generate different scope IDs")
	}
}

func TestInjectAssets(t *testing.T) {
	bundler := NewBundler("/out")

	html := `<html>
<head>
<title>Test</title>
</head>
<body>
<h1>Hello</h1>
</body>
</html>`

	result := bundler.InjectAssets(html, "/_assets/styles.css", "/_assets/script.js", "abc123")

	if !strings.Contains(result, `<link rel="stylesheet" href="/_assets/styles.css">`) {
		t.Error("Expected CSS link tag in head")
	}

	if !strings.Contains(result, `<script type="module" src="/_assets/script.js"></script>`) {
		t.Error("Expected script tag before closing body")
	}

	if !strings.Contains(result, `<body data-gxc-abc123>`) {
		t.Error("Expected scope attribute on body")
	}

	headIdx := strings.Index(result, "</head>")
	cssIdx := strings.Index(result, "styles.css")
	if cssIdx >= headIdx {
		t.Error("Expected CSS link before </head>")
	}

	bodyCloseIdx := strings.Index(result, "</body>")
	jsIdx := strings.Index(result, "script.js")
	if jsIdx >= bodyCloseIdx {
		t.Error("Expected script before </body>")
	}
}

func TestInjectAssetsNoCSS(t *testing.T) {
	bundler := NewBundler("/out")

	html := `<html><head></head><body></body></html>`
	result := bundler.InjectAssets(html, "", "/_assets/script.js", "")

	if strings.Contains(result, "<link") {
		t.Error("Expected no CSS link when cssPath is empty")
	}

	if !strings.Contains(result, "script.js") {
		t.Error("Expected script tag")
	}
}

func TestInjectAssetsNoJS(t *testing.T) {
	bundler := NewBundler("/out")

	html := `<html><head></head><body></body></html>`
	result := bundler.InjectAssets(html, "/_assets/styles.css", "", "")

	if !strings.Contains(result, "styles.css") {
		t.Error("Expected CSS link")
	}

	if strings.Contains(result, "<script") {
		t.Error("Expected no script tag when jsPath is empty")
	}
}

func TestInjectAssetsNoScope(t *testing.T) {
	bundler := NewBundler("/out")

	html := `<html><head></head><body></body></html>`
	result := bundler.InjectAssets(html, "", "", "")

	if strings.Contains(result, "data-gxc-") {
		t.Error("Expected no scope attribute when scopeID is empty")
	}
}

func TestBundleStylesHash(t *testing.T) {
	tmpDir := t.TempDir()
	bundler := NewBundler(tmpDir)

	comp1 := &parser.Component{
		Styles: []parser.Style{{Content: ".a { color: red; }", Scoped: false}},
	}

	comp2 := &parser.Component{
		Styles: []parser.Style{{Content: ".b { color: blue; }", Scoped: false}},
	}

	path1, _ := bundler.BundleStyles(comp1, "/page1.gxc")
	path2, _ := bundler.BundleStyles(comp2, "/page2.gxc")

	if path1 == path2 {
		t.Error("Expected different content to produce different hashes")
	}

	path3, _ := bundler.BundleStyles(comp1, "/page1.gxc")
	if path1 != path3 {
		t.Error("Expected same content to produce same hash")
	}
}

func TestBundleScriptsHash(t *testing.T) {
	tmpDir := t.TempDir()
	bundler := NewBundler(tmpDir)

	comp1 := &parser.Component{
		Scripts: []parser.Script{{Content: "console.log(1);", IsModule: false}},
	}

	comp2 := &parser.Component{
		Scripts: []parser.Script{{Content: "console.log(2);", IsModule: false}},
	}

	path1, _ := bundler.BundleScripts(comp1, "/page1.gxc")
	path2, _ := bundler.BundleScripts(comp2, "/page2.gxc")

	if path1 == path2 {
		t.Error("Expected different content to produce different hashes")
	}
}
