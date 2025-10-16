package wasm

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Compiler struct {
	TempDir   string
	CacheDir  string
	UseTinyGo bool
}

type CompiledModule struct {
	WasmPath   string
	LoaderPath string
	Hash       string
}

func NewCompiler(tempDir, cacheDir string) *Compiler {
	return &Compiler{
		TempDir:  tempDir,
		CacheDir: cacheDir,
	}
}

func (c *Compiler) Compile(script, pagePath string) (*CompiledModule, error) {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(script)))[:8]

	cachedWasm := filepath.Join(c.CacheDir, fmt.Sprintf("script-%s.wasm", hash))
	if _, err := os.Stat(cachedWasm); err == nil {
		return &CompiledModule{
			WasmPath: cachedWasm,
			Hash:     hash,
		}, nil
	}

	buildDir := filepath.Join(c.TempDir, hash)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return nil, fmt.Errorf("create build dir: %w", err)
	}
	defer os.RemoveAll(buildDir)

	preparedScript, err := prepareScript(script, hash, c.UseTinyGo && isTinyGoAvailable())
	if err != nil {
		return nil, fmt.Errorf("prepare script: %w", err)
	}

	mainGo := filepath.Join(buildDir, "main.go")
	if err := os.WriteFile(mainGo, []byte(preparedScript), 0644); err != nil {
		return nil, fmt.Errorf("write main.go: %w", err)
	}

	goMod := filepath.Join(buildDir, "go.mod")
	moduleContent := fmt.Sprintf("module wasmscript\n\ngo 1.21\n\nrequire github.com/galaxy/galaxy v0.0.0\n\nreplace github.com/galaxy/galaxy => %s\n", findModuleRoot())
	if err := os.WriteFile(goMod, []byte(moduleContent), 0644); err != nil {
		return nil, fmt.Errorf("write go.mod: %w", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = buildDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go mod tidy failed: %s\n%s", err, output)
	}

	outWasm := filepath.Join(buildDir, "script.wasm")

	var cmd *exec.Cmd
	if c.UseTinyGo && isTinyGoAvailable() {
		cmd = exec.Command("tinygo", "build", "-o", outWasm, "-target", "wasm", mainGo)
		cmd.Dir = buildDir
	} else {
		cmd = exec.Command("go", "build", "-o", outWasm, mainGo)
		cmd.Dir = buildDir
		cmd.Env = append(os.Environ(),
			"GOOS=js",
			"GOARCH=wasm",
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compile failed: %s\n%s", err, output)
	}

	if err := os.MkdirAll(c.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	finalWasm := filepath.Join(c.CacheDir, fmt.Sprintf("script-%s.wasm", hash))
	if err := os.Rename(outWasm, finalWasm); err != nil {
		data, err := os.ReadFile(outWasm)
		if err != nil {
			return nil, fmt.Errorf("read wasm: %w", err)
		}
		if err := os.WriteFile(finalWasm, data, 0644); err != nil {
			return nil, fmt.Errorf("write cached wasm: %w", err)
		}
	}

	return &CompiledModule{
		WasmPath: finalWasm,
		Hash:     hash,
	}, nil
}

func findModuleRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func prepareScript(script, hash string, useTinyGo bool) (string, error) {
	imports := extractImports(script)
	body := removeImports(script)
	body = removePackageDecl(body)
	hasMain := containsMainFunc(body)

	pkgName := "main"
	if !useTinyGo {
		pkgName = fmt.Sprintf("wasmscript_%s", hash[:8])
	}

	var final strings.Builder
	final.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	if len(imports) > 0 {
		final.WriteString("import (\n")
		for _, imp := range imports {
			final.WriteString(fmt.Sprintf("\t%s\n", imp))
		}
		final.WriteString(")\n\n")
	}

	if !hasMain {
		final.WriteString("func main() {\n")
		final.WriteString(indentCode(body))
		final.WriteString("\n\tselect {}\n")
		final.WriteString("}\n")
	} else {
		final.WriteString(body)
	}

	return final.String(), nil
}

func extractImports(script string) []string {
	var imports []string
	importRegex := regexp.MustCompile(`(?m)^import\s+(.+)$`)
	matches := importRegex.FindAllStringSubmatch(script, -1)

	for _, match := range matches {
		imports = append(imports, match[1])
	}

	return imports
}

func removeImports(script string) string {
	importRegex := regexp.MustCompile(`(?m)^import\s+.+$\n?`)
	return importRegex.ReplaceAllString(script, "")
}

func removePackageDecl(script string) string {
	pkgRegex := regexp.MustCompile(`(?m)^package\s+\w+\s*\n?`)
	return pkgRegex.ReplaceAllString(script, "")
}

func containsMainFunc(body string) bool {
	fset := token.NewFileSet()
	wrapped := "package temp\n" + body

	f, err := parser.ParseFile(fset, "", wrapped, 0)
	if err != nil {
		return false
	}

	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == "main" {
				return true
			}
		}
	}

	return false
}

func indentCode(code string) string {
	lines := strings.Split(strings.TrimSpace(code), "\n")
	var result strings.Builder

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		if strings.TrimSpace(line) != "" {
			result.WriteString("\t")
		}
		result.WriteString(line)
	}

	return result.String()
}

func isTinyGoAvailable() bool {
	_, err := exec.LookPath("tinygo")
	return err == nil
}
