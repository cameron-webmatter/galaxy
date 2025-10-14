package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gastro/gastro/pkg/build"
	"github.com/gastro/gastro/pkg/executor"
	"github.com/gastro/gastro/pkg/parser"
	"github.com/gastro/gastro/pkg/server"
	"github.com/gastro/gastro/pkg/template"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "dev":
		runDev()
	case "build":
		runBuild()
	case "test":
		runTest()
	case "lsp-server":
		runLSP()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Gastro - Go-powered web framework")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gastro dev        Start development server")
	fmt.Println("  gastro build      Build for production")
	fmt.Println("  gastro test       Run framework tests")
	fmt.Println("  gastro lsp-server Start LSP server (--stdio)")
}

func runLSP() {
	os.Args = append([]string{"lsp-server"}, os.Args[2:]...)
	execLSP()
}

func execLSP() {
	fmt.Fprintln(os.Stderr, "LSP server started. Use with VS Code extension.")
}

func runDev() {
	pagesDir := "./pages"
	publicDir := "./public"
	port := 3000

	if len(os.Args) > 2 {
		pagesDir = os.Args[2]
	}

	srv := server.NewDevServer(pagesDir, publicDir, port)

	if err := srv.Start(); err != nil {
		log.Fatalf("Dev server error: %v", err)
	}
}

func runBuild() {
	fmt.Println("Building Gastro project...")

	pagesDir := "./pages"
	outDir := "./dist"
	publicDir := "./public"

	if len(os.Args) > 2 {
		pagesDir = os.Args[2]
	}
	if len(os.Args) > 3 {
		outDir = os.Args[3]
	}

	builder := build.NewSSGBuilder(pagesDir, outDir, publicDir)

	if err := builder.Build(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}

	fmt.Printf("\n✓ Build complete! Output: %s\n", outDir)
}

func runTest() {
	fmt.Println("Testing Gastro components...")

	testFile := "./examples/basic/pages/index.gxc"
	if len(os.Args) > 2 {
		testFile = os.Args[2]
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		log.Fatalf("Failed to parse component: %v", err)
	}

	fmt.Println(comp.String())

	ctx := executor.NewContext()
	if err := ctx.Execute(comp.Frontmatter); err != nil {
		log.Fatalf("Failed to execute frontmatter: %v", err)
	}

	fmt.Println(ctx.String())

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(comp.Template, nil)
	if err != nil {
		log.Fatalf("Failed to render template: %v", err)
	}

	fmt.Println("\n=== Rendered Output ===")
	fmt.Println(rendered)
}

func renderPage(filePath string, params map[string]string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}

	comp, err := parser.Parse(string(content))
	if err != nil {
		log.Printf("Error parsing component: %v", err)
		return
	}

	ctx := executor.NewContext()

	for k, v := range params {
		ctx.Set(k, v)
	}

	if err := ctx.Execute(comp.Frontmatter); err != nil {
		log.Printf("Error executing frontmatter: %v", err)
		return
	}

	engine := template.NewEngine(ctx)
	rendered, err := engine.Render(comp.Template, nil)
	if err != nil {
		log.Printf("Error rendering: %v", err)
		return
	}

	fmt.Println("\n=== Rendered Page ===")
	fmt.Println(rendered)
}
