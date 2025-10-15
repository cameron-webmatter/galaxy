package template

import (
	"strings"
	"testing"

	"github.com/galaxy/galaxy/pkg/executor"
)

func TestNestedTagsInForDirective(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var posts = []map[string]string{
		{"title": "First Post", "slug": "first-post"},
		{"title": "Second Post", "slug": "second-post"},
	}`)

	engine := NewEngine(ctx)
	template := `<article galaxy:for={post in posts}><h2>{post.title}</h2><a href="/blog/{post.slug}">Read</a></article>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, "<h2>First Post</h2>") {
		t.Errorf("expected h2 with First Post, got: %s", result)
	}
	if !strings.Contains(result, `<a href="/blog/first-post">Read</a>`) {
		t.Errorf("expected link to first-post, got: %s", result)
	}
	if !strings.Contains(result, "<h2>Second Post</h2>") {
		t.Errorf("expected h2 with Second Post, got: %s", result)
	}
}

func TestDeeplyNestedTags(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var items = []map[string]string{
		{"name": "Item A", "price": "10"},
	}`)

	engine := NewEngine(ctx)
	template := `<div galaxy:for={item in items}><section><article><h3>{item.name}</h3><p>${item.price}</p></article></section></div>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, "<h3>Item A</h3>") {
		t.Errorf("expected h3, got: %s", result)
	}
	if !strings.Contains(result, "<p>$10</p>") {
		t.Errorf("expected p tag, got: %s", result)
	}
}

func TestSameTagNested(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var items = []string{"A", "B"}`)

	engine := NewEngine(ctx)
	template := `<div galaxy:for={item in items}><div class="inner">{item}</div></div>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, `<div class="inner">A</div>`) {
		t.Errorf("expected inner div A, got: %s", result)
	}
	if !strings.Contains(result, `<div class="inner">B</div>`) {
		t.Errorf("expected inner div B, got: %s", result)
	}
}

func TestMultipleDirectivesInTemplate(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var posts = []string{"Post 1", "Post 2"}
var comments = []string{"Comment 1"}`)

	engine := NewEngine(ctx)
	template := `<div><li galaxy:for={post in posts}>{post}</li><span galaxy:for={comment in comments}>{comment}</span></div>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, "<li>Post 1</li>") {
		t.Errorf("expected Post 1, got: %s", result)
	}
	if !strings.Contains(result, "<li>Post 2</li>") {
		t.Errorf("expected Post 2, got: %s", result)
	}
	if !strings.Contains(result, "<span>Comment 1</span>") {
		t.Errorf("expected Comment 1, got: %s", result)
	}
}

func TestNestedIfDirective(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var show = 1`)

	engine := NewEngine(ctx)
	template := `<div galaxy:if={show}><h1>Title</h1><p>Content</p></div>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, "<h1>Title</h1>") {
		t.Errorf("expected h1, got: %s", result)
	}
	if !strings.Contains(result, "<p>Content</p>") {
		t.Errorf("expected p, got: %s", result)
	}
}

func TestForWithAdditionalAttributes(t *testing.T) {
	ctx := executor.NewContext()
	ctx.Execute(`var items = []string{"A", "B"}`)

	engine := NewEngine(ctx)
	template := `<div galaxy:for={item in items} class="item" data-id="123"><span>{item}</span></div>`

	result, err := engine.Render(template, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, `class="item"`) {
		t.Errorf("expected class attribute, got: %s", result)
	}
	if !strings.Contains(result, `data-id="123"`) {
		t.Errorf("expected data-id attribute, got: %s", result)
	}
	if !strings.Contains(result, "<span>A</span>") {
		t.Errorf("expected span A, got: %s", result)
	}
}
