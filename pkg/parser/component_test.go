package parser

import (
	"testing"
)

func TestParseBasicComponent(t *testing.T) {
	input := `---
title := "Hello World"
count := 42
---
<h1>{title}</h1>
<p>Count: {count}</p>
`

	comp, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if !containsString(comp.Frontmatter, "title") {
		t.Error("Frontmatter should contain title")
	}

	if !containsString(comp.Template, "<h1>") {
		t.Error("Template should contain h1 tag")
	}
}

func TestParseComponentWithScripts(t *testing.T) {
	input := `---
data := "test"
---
<div id="app"></div>

<script>
console.log("Hello");
</script>

<script type="module">
import { foo } from './bar.js';
</script>
`

	comp, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(comp.Scripts) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(comp.Scripts))
	}

	if !comp.Scripts[1].IsModule {
		t.Error("Second script should be module")
	}
}

func TestParseComponentWithStyles(t *testing.T) {
	input := `---
---
<div class="container"></div>

<style scoped>
.container { color: red; }
</style>

<style>
body { margin: 0; }
</style>
`

	comp, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(comp.Styles) != 2 {
		t.Errorf("Expected 2 styles, got %d", len(comp.Styles))
	}

	if !comp.Styles[0].Scoped {
		t.Error("First style should be scoped")
	}

	if comp.Styles[1].Scoped {
		t.Error("Second style should not be scoped")
	}
}

func TestParseImports(t *testing.T) {
	input := `---
import Layout from "../layouts/Base.gxc"
import Card from "../components/Card.gxc"

title := "My Page"
---
<Layout title={title}>
  <Card />
</Layout>
`

	comp, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(comp.Imports) != 2 {
		t.Errorf("Expected 2 imports, got %d", len(comp.Imports))
	}

	if comp.Imports[0].Alias != "Layout" {
		t.Errorf("Expected alias 'Layout', got '%s'", comp.Imports[0].Alias)
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) && s[:len(substr)] == substr ||
			len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
			len(s) > len(substr) && s[0] != substr[0] && s[len(s)-1] != substr[len(substr)-1])
}
