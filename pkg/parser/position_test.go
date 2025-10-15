package parser

import (
	"testing"
)

func TestNewRange(t *testing.T) {
	r := NewRange(1, 5, 3, 10)

	if r.Start.Line != 1 || r.Start.Column != 5 {
		t.Errorf("Expected start (1,5), got (%d,%d)", r.Start.Line, r.Start.Column)
	}

	if r.End.Line != 3 || r.End.Column != 10 {
		t.Errorf("Expected end (3,10), got (%d,%d)", r.End.Line, r.End.Column)
	}
}

func TestLineColFromOffset(t *testing.T) {
	tests := []struct {
		name    string
		content string
		offset  int
		wantL   int
		wantC   int
	}{
		{"empty string", "", 0, 1, 1},
		{"start of string", "hello", 0, 1, 1},
		{"middle of first line", "hello world", 6, 1, 7},
		{"after newline", "hello\nworld", 6, 2, 1},
		{"multiple newlines", "a\nb\nc", 4, 3, 1},
		{"second char second line", "foo\nbar", 5, 2, 2},
		{"out of bounds", "test", 100, 1, 5},
		{"single char", "x", 0, 1, 1},
		{"newline only", "\n", 0, 1, 1},
		{"after single newline", "\n", 1, 2, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotL, gotC := lineColFromOffset(tt.content, tt.offset)
			if gotL != tt.wantL || gotC != tt.wantC {
				t.Errorf("lineColFromOffset(%q, %d) = (%d,%d), want (%d,%d)",
					tt.content, tt.offset, gotL, gotC, tt.wantL, tt.wantC)
			}
		})
	}
}

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		token TokenType
		want  string
	}{
		{TokenFrontmatter, "Frontmatter"},
		{TokenTemplate, "Template"},
		{TokenScript, "Script"},
		{TokenStyle, "Style"},
		{TokenDirective, "Directive"},
		{TokenExpression, "Expression"},
		{TokenHTMLTag, "HTMLTag"},
		{TokenComment, "Comment"},
		{TokenType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.token.String()
			if got != tt.want {
				t.Errorf("TokenType(%d).String() = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func TestTokenCreation(t *testing.T) {
	tok := Token{
		Type:  TokenScript,
		Value: "console.log('test')",
		Range: NewRange(5, 1, 5, 20),
	}

	if tok.Type != TokenScript {
		t.Errorf("Expected TokenScript, got %v", tok.Type)
	}
	if tok.Value != "console.log('test')" {
		t.Errorf("Expected value mismatch")
	}
}

func TestPositionStruct(t *testing.T) {
	p := Position{Line: 10, Column: 25}

	if p.Line != 10 {
		t.Errorf("Expected line 10, got %d", p.Line)
	}
	if p.Column != 25 {
		t.Errorf("Expected column 25, got %d", p.Column)
	}
}
