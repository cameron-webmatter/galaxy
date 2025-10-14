package router

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStaticRoute(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "index.gxc"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "about.gxc"), []byte(""), 0644)

	router := NewRouter(tmpDir)
	err := router.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	router.Sort()

	if len(router.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(router.Routes))
	}

	route, params := router.Match("/")
	if route == nil {
		t.Error("Expected to match /")
	}
	if len(params) != 0 {
		t.Errorf("Expected no params, got %d", len(params))
	}

	route, params = router.Match("/about")
	if route == nil {
		t.Error("Expected to match /about")
	}
}

func TestDynamicRoute(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "posts"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "posts", "[id].gxc"), []byte(""), 0644)

	router := NewRouter(tmpDir)
	err := router.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	router.Sort()

	route, params := router.Match("/posts/123")
	if route == nil {
		t.Fatal("Expected to match /posts/123")
	}

	if route.Type != RouteDynamic {
		t.Error("Expected dynamic route")
	}

	if params["id"] != "123" {
		t.Errorf("Expected id=123, got %s", params["id"])
	}
}

func TestCatchAllRoute(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "[...slug].gxc"), []byte(""), 0644)

	router := NewRouter(tmpDir)
	err := router.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	router.Sort()

	route, params := router.Match("/any/path/here")
	if route == nil {
		t.Fatal("Expected to match /any/path/here")
	}

	if route.Type != RouteCatchAll {
		t.Error("Expected catch-all route")
	}

	if params["slug"] != "any/path/here" {
		t.Errorf("Expected slug=any/path/here, got %s", params["slug"])
	}
}

func TestRoutePriority(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "blog"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "blog", "featured.gxc"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "blog", "[slug].gxc"), []byte(""), 0644)

	router := NewRouter(tmpDir)
	err := router.Discover()
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	router.Sort()

	route, _ := router.Match("/blog/featured")
	if route == nil {
		t.Fatal("Expected to match /blog/featured")
	}

	if route.Type != RouteStatic {
		t.Error("Expected static route to have higher priority")
	}
}
