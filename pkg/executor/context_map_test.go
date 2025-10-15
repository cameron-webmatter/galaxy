package executor

import (
	"testing"
)

func TestMapLiteral(t *testing.T) {
	ctx := NewContext()

	code := `var user = map[string]string{"name": "John", "email": "john@example.com"}`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	val, ok := ctx.Get("user")
	if !ok {
		t.Fatal("user variable not found")
	}

	userMap, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", val)
	}

	if userMap["name"] != "John" {
		t.Errorf("expected name='John', got %v", userMap["name"])
	}

	if userMap["email"] != "john@example.com" {
		t.Errorf("expected email='john@example.com', got %v", userMap["email"])
	}
}

func TestSliceOfMaps(t *testing.T) {
	ctx := NewContext()

	code := `var posts = []map[string]string{
		{"title": "First Post", "slug": "first-post"},
		{"title": "Second Post", "slug": "second-post"},
	}`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	val, ok := ctx.Get("posts")
	if !ok {
		t.Fatal("posts variable not found")
	}

	posts, ok := val.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", val)
	}

	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	post1, ok := posts[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", posts[0])
	}

	if post1["title"] != "First Post" {
		t.Errorf("expected title='First Post', got %v", post1["title"])
	}

	if post1["slug"] != "first-post" {
		t.Errorf("expected slug='first-post', got %v", post1["slug"])
	}

	post2, ok := posts[1].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", posts[1])
	}

	if post2["title"] != "Second Post" {
		t.Errorf("expected title='Second Post', got %v", post2["title"])
	}
}

func TestMapWithIntValues(t *testing.T) {
	ctx := NewContext()

	code := `var scores = map[string]int{"math": 95, "english": 88}`

	err := ctx.Execute(code)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	val, ok := ctx.Get("scores")
	if !ok {
		t.Fatal("scores variable not found")
	}

	scores, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", val)
	}

	if scores["math"] != int64(95) {
		t.Errorf("expected math=95, got %v", scores["math"])
	}

	if scores["english"] != int64(88) {
		t.Errorf("expected english=88, got %v", scores["english"])
	}
}
