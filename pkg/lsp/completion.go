package lsp

import (
	"github.com/cameron-webmatter/galaxy/pkg/executor"
	"github.com/cameron-webmatter/galaxy/pkg/parser"
	"go.lsp.dev/protocol"
)

func (s *Server) getCompletions(content string, pos protocol.Position) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0)

	comp, err := parser.Parse(content)
	if err != nil {
		return items
	}

	if comp.Frontmatter != "" {
		ctx := executor.NewContext()
		ctx.Execute(comp.Frontmatter)

		for varName := range ctx.Variables {
			items = append(items, protocol.CompletionItem{
				Label:  varName,
				Kind:   protocol.CompletionItemKindVariable,
				Detail: "Variable from frontmatter",
			})
		}
	}

	directiveCompletions := []protocol.CompletionItem{
		{
			Label:  "galaxy:if",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: "Conditional rendering",
		},
		{
			Label:  "galaxy:elsif",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: "Else-if conditional branch",
		},
		{
			Label:  "galaxy:else",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: "Else conditional branch",
		},
		{
			Label:  "galaxy:for",
			Kind:   protocol.CompletionItemKindKeyword,
			Detail: "Loop rendering",
		},
	}
	items = append(items, directiveCompletions...)

	return items
}
