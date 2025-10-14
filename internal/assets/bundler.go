package assets

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gastro/gastro/pkg/parser"
)

type Bundler struct {
	OutDir string
}

func NewBundler(outDir string) *Bundler {
	return &Bundler{OutDir: outDir}
}

func (b *Bundler) BundleStyles(comp *parser.Component, pagePath string) (string, error) {
	if len(comp.Styles) == 0 {
		return "", nil
	}

	var combined strings.Builder
	for _, style := range comp.Styles {
		if style.Scoped {
			scopedCSS := b.scopeCSS(style.Content, pagePath)
			combined.WriteString(scopedCSS)
		} else {
			combined.WriteString(style.Content)
		}
		combined.WriteString("\n")
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(combined.String())))[:8]
	filename := fmt.Sprintf("styles-%s.css", hash)
	outPath := filepath.Join(b.OutDir, "_assets", filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(outPath, []byte(combined.String()), 0644); err != nil {
		return "", err
	}

	return "/_assets/" + filename, nil
}

func (b *Bundler) BundleScripts(comp *parser.Component, pagePath string) (string, error) {
	if len(comp.Scripts) == 0 {
		return "", nil
	}

	var combined strings.Builder
	for i, script := range comp.Scripts {
		if i > 0 {
			combined.WriteString("\n")
		}
		combined.WriteString(script.Content)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(combined.String())))[:8]
	filename := fmt.Sprintf("script-%s.js", hash)
	outPath := filepath.Join(b.OutDir, "_assets", filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(outPath, []byte(combined.String()), 0644); err != nil {
		return "", err
	}

	return "/_assets/" + filename, nil
}

func (b *Bundler) scopeCSS(css, pagePath string) string {
	scope := b.GenerateScopeID(pagePath)

	lines := strings.Split(css, "\n")
	var scoped strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "/*") {
			scoped.WriteString(line)
			scoped.WriteString("\n")
			continue
		}

		if strings.Contains(trimmed, "{") {
			parts := strings.SplitN(trimmed, "{", 2)
			selector := strings.TrimSpace(parts[0])
			rest := parts[1]

			scoped.WriteString(fmt.Sprintf("[data-gxc-%s] %s { %s\n", scope, selector, rest))
		} else {
			scoped.WriteString(line)
			scoped.WriteString("\n")
		}
	}

	return scoped.String()
}

func (b *Bundler) GenerateScopeID(pagePath string) string {
	hash := sha256.Sum256([]byte(pagePath))
	return fmt.Sprintf("%x", hash)[:6]
}

func (b *Bundler) InjectAssets(html, cssPath, jsPath, scopeID string) string {
	if scopeID != "" {
		bodyScopeAttr := fmt.Sprintf(`data-gxc-%s`, scopeID)
		html = strings.Replace(html, "<body>", fmt.Sprintf(`<body %s>`, bodyScopeAttr), 1)
	}

	if cssPath != "" {
		cssTag := fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssPath)
		html = strings.Replace(html, "</head>", cssTag+"\n</head>", 1)
	}

	if jsPath != "" {
		jsTag := fmt.Sprintf(`<script type="module" src="%s"></script>`, jsPath)
		html = strings.Replace(html, "</body>", jsTag+"\n</body>", 1)
	}

	return html
}
