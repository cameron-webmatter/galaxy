package lsp

import (
	"fmt"

	"github.com/galaxy/galaxy/pkg/executor"
	"github.com/galaxy/galaxy/pkg/parser"
	"go.lsp.dev/protocol"
)

func (s *Server) getHover(content string, pos protocol.Position) *protocol.Hover {
	comp, err := parser.Parse(content)
	if err != nil {
		return nil
	}

	if comp.Frontmatter == "" {
		return nil
	}

	ctx := executor.NewContext()
	ctx.Execute(comp.Frontmatter)

	for varName, value := range ctx.Variables {
		hoverText := fmt.Sprintf("**%s**: `%T`\n\nValue: `%v`", varName, value, value)

		return &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: hoverText,
			},
		}
	}

	return nil
}
