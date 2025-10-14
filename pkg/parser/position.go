package parser

type Position struct {
	Line   int
	Column int
}

type Range struct {
	Start Position
	End   Position
}

type Token struct {
	Type  TokenType
	Value string
	Range Range
}

type TokenType int

const (
	TokenFrontmatter TokenType = iota
	TokenTemplate
	TokenScript
	TokenStyle
	TokenDirective
	TokenExpression
	TokenHTMLTag
	TokenComment
)

func (t TokenType) String() string {
	names := []string{
		"Frontmatter",
		"Template",
		"Script",
		"Style",
		"Directive",
		"Expression",
		"HTMLTag",
		"Comment",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

func NewRange(startLine, startCol, endLine, endCol int) Range {
	return Range{
		Start: Position{Line: startLine, Column: startCol},
		End:   Position{Line: endLine, Column: endCol},
	}
}

func lineColFromOffset(content string, offset int) (int, int) {
	line := 1
	col := 1

	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}

	return line, col
}
