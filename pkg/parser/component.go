package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Component struct {
	Frontmatter      string
	FrontmatterRange Range
	Template         string
	TemplateRange    Range
	Scripts          []Script
	Styles           []Style
	Imports          []Import
	Tokens           []Token
	Expressions      []Expression
	Directives       []Directive
}

type Expression struct {
	Content   string
	Range     Range
	Variables []string
}

type Directive struct {
	Name      string
	Condition string
	Range     Range
}

type Script struct {
	Content  string
	IsModule bool
}

type Style struct {
	Content string
	Scoped  bool
}

type Import struct {
	Path      string
	Alias     string
	IsDefault bool
}

var (
	frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n?`)
	scriptRegex      = regexp.MustCompile(`(?s)<script(?:\s+([^>]*))?>(.+?)</script>`)
	styleRegex       = regexp.MustCompile(`(?s)<style(?:\s+([^>]*))?>(.+?)</style>`)
	importRegex      = regexp.MustCompile(`import\s+(?:(\w+)\s+from\s+)?['"](.*?)['"]`)
)

func Parse(content string) (*Component, error) {
	comp := &Component{
		Tokens:      make([]Token, 0),
		Expressions: make([]Expression, 0),
		Directives:  make([]Directive, 0),
	}

	originalContent := content

	frontmatterMatch := frontmatterRegex.FindStringSubmatchIndex(content)
	if frontmatterMatch != nil {
		comp.Frontmatter = strings.TrimSpace(content[frontmatterMatch[2]:frontmatterMatch[3]])

		startLine, startCol := lineColFromOffset(originalContent, frontmatterMatch[0])
		endLine, endCol := lineColFromOffset(originalContent, frontmatterMatch[1])
		comp.FrontmatterRange = NewRange(startLine, startCol, endLine, endCol)

		comp.Tokens = append(comp.Tokens, Token{
			Type:  TokenFrontmatter,
			Value: comp.Frontmatter,
			Range: comp.FrontmatterRange,
		})

		content = frontmatterRegex.ReplaceAllString(content, "")

		comp.Imports = parseImports(comp.Frontmatter)
	}

	scriptMatches := scriptRegex.FindAllStringSubmatch(content, -1)
	for _, match := range scriptMatches {
		attrs := match[1]
		isModule := strings.Contains(attrs, "type=\"module\"") || strings.Contains(attrs, `type='module'`)
		comp.Scripts = append(comp.Scripts, Script{
			Content:  strings.TrimSpace(match[2]),
			IsModule: isModule,
		})
	}
	content = scriptRegex.ReplaceAllString(content, "")

	styleMatches := styleRegex.FindAllStringSubmatch(content, -1)
	for _, match := range styleMatches {
		attrs := match[1]
		scoped := strings.Contains(attrs, "scoped")
		comp.Styles = append(comp.Styles, Style{
			Content: strings.TrimSpace(match[2]),
			Scoped:  scoped,
		})
	}
	content = styleRegex.ReplaceAllString(content, "")

	comp.Template = strings.TrimSpace(content)

	return comp, nil
}

func parseImports(frontmatter string) []Import {
	var imports []Import
	lines := strings.Split(frontmatter, "\n")

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "import") {
			matches := importRegex.FindStringSubmatch(line)
			if matches != nil {
				imp := Import{
					Path:      matches[2],
					IsDefault: matches[1] != "",
				}
				if matches[1] != "" {
					imp.Alias = matches[1]
				}
				imports = append(imports, imp)
			}
		}
	}

	return imports
}

func (c *Component) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "=== Component ===\n")
	fmt.Fprintf(&buf, "Frontmatter:\n%s\n\n", c.Frontmatter)
	fmt.Fprintf(&buf, "Template:\n%s\n\n", c.Template)
	fmt.Fprintf(&buf, "Scripts: %d\n", len(c.Scripts))
	fmt.Fprintf(&buf, "Styles: %d\n", len(c.Styles))
	fmt.Fprintf(&buf, "Imports: %d\n", len(c.Imports))

	return buf.String()
}
