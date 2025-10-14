package executor

import (
	"testing"
)

func TestExecuteSimpleAssignment(t *testing.T) {
	ctx := NewContext()

	code := `
var title = "Hello"
var count = 42
var pi = 3.14
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("title"); !ok || val != "Hello" {
		t.Errorf("Expected title='Hello', got %v", val)
	}

	if val, ok := ctx.Get("count"); !ok || val != int64(42) {
		t.Errorf("Expected count=42, got %v", val)
	}

	if val, ok := ctx.Get("pi"); !ok || val != 3.14 {
		t.Errorf("Expected pi=3.14, got %v", val)
	}
}

func TestExecuteArithmetic(t *testing.T) {
	ctx := NewContext()

	code := `
var a = 10
var b = 5
var sum = a + b
var diff = a - b
var prod = a * b
var quot = a / b
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	tests := []struct {
		name     string
		expected int64
	}{
		{"sum", 15},
		{"diff", 5},
		{"prod", 50},
		{"quot", 2},
	}

	for _, tt := range tests {
		if val, ok := ctx.Get(tt.name); !ok || val != tt.expected {
			t.Errorf("Expected %s=%d, got %v", tt.name, tt.expected, val)
		}
	}
}

func TestExecuteStringConcat(t *testing.T) {
	ctx := NewContext()

	code := `
var first = "Hello"
var last = "World"
var greeting = first + " " + last
`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if val, ok := ctx.Get("greeting"); !ok || val != "Hello World" {
		t.Errorf("Expected greeting='Hello World', got %v", val)
	}
}

func TestPropsAndSlots(t *testing.T) {
	ctx := NewContext()

	ctx.SetProp("title", "Page Title")
	ctx.SetProp("count", 42)

	if val, ok := ctx.GetProp("title"); !ok || val != "Page Title" {
		t.Errorf("Expected prop title='Page Title', got %v", val)
	}

	if val, ok := ctx.GetProp("count"); !ok || val != 42 {
		t.Errorf("Expected prop count=42, got %v", val)
	}
}
