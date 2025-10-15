package ssr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Island struct {
	ComponentPath string
	Props         map[string]interface{}
	Strategy      string
	ID            string
}

func NewIsland(componentPath string, props map[string]interface{}, strategy string) *Island {
	id := generateIslandID(componentPath, props)
	return &Island{
		ComponentPath: componentPath,
		Props:         props,
		Strategy:      strategy,
		ID:            id,
	}
}

func generateIslandID(path string, props map[string]interface{}) string {
	data := fmt.Sprintf("%s:%v", path, props)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:12]
}

func (i *Island) RenderScript() string {
	propsJSON, _ := json.Marshal(i.Props)

	return fmt.Sprintf(
		`<script type="module">
  import { hydrate } from '/_galaxy/hydration.js';
  hydrate('%s', '%s', %s, '%s');
</script>`,
		i.ID,
		i.ComponentPath,
		string(propsJSON),
		i.Strategy,
	)
}

func (i *Island) WrapContent(content string) string {
	return fmt.Sprintf(
		`<div data-island-id="%s" data-island-strategy="%s">%s</div>%s`,
		i.ID,
		i.Strategy,
		content,
		i.RenderScript(),
	)
}
