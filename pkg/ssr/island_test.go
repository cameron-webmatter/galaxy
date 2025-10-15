package ssr

import (
	"strings"
	"testing"
)

func TestNewIsland(t *testing.T) {
	props := map[string]interface{}{
		"title": "Hello",
		"count": 42,
	}

	island := NewIsland("/components/Counter.tsx", props, "load")

	if island.ComponentPath != "/components/Counter.tsx" {
		t.Errorf("Expected ComponentPath /components/Counter.tsx, got %s", island.ComponentPath)
	}

	if island.Strategy != "load" {
		t.Errorf("Expected strategy load, got %s", island.Strategy)
	}

	if island.ID == "" {
		t.Error("Expected ID to be generated")
	}

	if len(island.ID) != 12 {
		t.Errorf("Expected ID length 12, got %d", len(island.ID))
	}
}

func TestGenerateIslandIDDeterministic(t *testing.T) {
	props := map[string]interface{}{
		"name": "test",
		"val":  123,
	}

	id1 := generateIslandID("/comp/A.tsx", props)
	id2 := generateIslandID("/comp/A.tsx", props)

	if id1 != id2 {
		t.Error("Expected same props to generate same ID")
	}
}

func TestGenerateIslandIDUnique(t *testing.T) {
	props1 := map[string]interface{}{"a": 1}
	props2 := map[string]interface{}{"a": 2}

	id1 := generateIslandID("/comp/A.tsx", props1)
	id2 := generateIslandID("/comp/A.tsx", props2)

	if id1 == id2 {
		t.Error("Expected different props to generate different IDs")
	}

	id3 := generateIslandID("/comp/B.tsx", props1)
	if id1 == id3 {
		t.Error("Expected different paths to generate different IDs")
	}
}

func TestIslandRenderScript(t *testing.T) {
	props := map[string]interface{}{
		"title":   "Test",
		"visible": true,
	}

	island := NewIsland("/components/Card.tsx", props, "idle")
	script := island.RenderScript()

	if !strings.Contains(script, "<script type=\"module\">") {
		t.Error("Expected module script tag")
	}

	if !strings.Contains(script, "import { hydrate }") {
		t.Error("Expected hydrate import")
	}

	if !strings.Contains(script, island.ID) {
		t.Error("Expected script to contain island ID")
	}

	if !strings.Contains(script, "/components/Card.tsx") {
		t.Error("Expected script to contain component path")
	}

	if !strings.Contains(script, "idle") {
		t.Error("Expected script to contain strategy")
	}

	if !strings.Contains(script, "\"title\":\"Test\"") {
		t.Error("Expected script to contain serialized props")
	}
}

func TestIslandRenderScriptEmptyProps(t *testing.T) {
	island := NewIsland("/components/Simple.tsx", map[string]interface{}{}, "load")
	script := island.RenderScript()

	if !strings.Contains(script, "{}") {
		t.Error("Expected empty props object in script")
	}
}

func TestIslandWrapContent(t *testing.T) {
	props := map[string]interface{}{"name": "Test"}
	island := NewIsland("/components/Widget.tsx", props, "visible")

	content := "<div>Widget Content</div>"
	wrapped := island.WrapContent(content)

	if !strings.Contains(wrapped, content) {
		t.Error("Expected wrapped content to contain original content")
	}

	if !strings.Contains(wrapped, "data-island-id=\""+island.ID+"\"") {
		t.Error("Expected data-island-id attribute")
	}

	if !strings.Contains(wrapped, "data-island-strategy=\"visible\"") {
		t.Error("Expected data-island-strategy attribute")
	}

	if !strings.Contains(wrapped, "<script type=\"module\">") {
		t.Error("Expected script to be included in wrapped content")
	}
}

func TestIslandWrapContentStructure(t *testing.T) {
	island := NewIsland("/comp/A.tsx", map[string]interface{}{}, "load")
	wrapped := island.WrapContent("<p>Test</p>")

	if !strings.HasPrefix(wrapped, "<div data-island-id=") {
		t.Error("Expected wrapped content to start with div")
	}

	if !strings.Contains(wrapped, "</div>") {
		t.Error("Expected closing div tag")
	}

	divCloseIdx := strings.Index(wrapped, "</div>")
	scriptIdx := strings.Index(wrapped, "<script")

	if scriptIdx <= divCloseIdx {
		t.Error("Expected script to come after closing div")
	}
}

func TestMultipleIslandsUnique(t *testing.T) {
	island1 := NewIsland("/comp/A.tsx", map[string]interface{}{"id": 1}, "load")
	island2 := NewIsland("/comp/A.tsx", map[string]interface{}{"id": 2}, "load")
	island3 := NewIsland("/comp/B.tsx", map[string]interface{}{"id": 1}, "load")

	ids := map[string]bool{
		island1.ID: true,
		island2.ID: true,
		island3.ID: true,
	}

	if len(ids) != 3 {
		t.Error("Expected 3 unique island IDs")
	}
}

func TestIslandStrategies(t *testing.T) {
	strategies := []string{"load", "idle", "visible", "media", "only"}

	for _, strategy := range strategies {
		island := NewIsland("/comp/Test.tsx", map[string]interface{}{}, strategy)

		if island.Strategy != strategy {
			t.Errorf("Expected strategy %s, got %s", strategy, island.Strategy)
		}

		script := island.RenderScript()
		if !strings.Contains(script, "'"+strategy+"'") {
			t.Errorf("Expected strategy %s in script", strategy)
		}
	}
}
