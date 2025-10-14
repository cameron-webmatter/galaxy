package lsp

import (
	"github.com/gastro/gastro/pkg/executor"
	"github.com/gastro/gastro/pkg/parser"
	"go.lsp.dev/protocol"
)

func (s *Server) analyze(content string) []protocol.Diagnostic {
	diagnostics := make([]protocol.Diagnostic, 0)

	comp, err := parser.Parse(content)
	if err != nil {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 1},
			},
			Severity: protocol.DiagnosticSeverityError,
			Source:   "gxc-parser",
			Message:  err.Error(),
		})
		return diagnostics
	}

	if comp.Frontmatter != "" {
		ctx := executor.NewContext()
		if err := ctx.Execute(comp.Frontmatter); err != nil {
			r := comp.FrontmatterRange
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(r.Start.Line - 1), Character: uint32(r.Start.Column)},
					End:   protocol.Position{Line: uint32(r.End.Line - 1), Character: uint32(r.End.Column)},
				},
				Severity: protocol.DiagnosticSeverityError,
				Source:   "gxc-executor",
				Message:  err.Error(),
			})
		}
	}

	return diagnostics
}
