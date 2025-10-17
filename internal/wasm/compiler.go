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
	moduleRoot := findModuleRoot()
	if moduleRoot == "" {
		return nil, fmt.Errorf("could not find galaxy module root")
	}
	moduleContent := fmt.Sprintf("module wasmscript\n\ngo 1.21\n\nrequire github.com/cameron-webmatter/galaxy v0.0.0\n\nreplace github.com/cameron-webmatter/galaxy => %s\n", moduleRoot)
	if err := os.WriteFile(goMod, []byte(moduleContent), 0644); err != nil {
		return nil, fmt.Errorf("write go.mod: %w", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = buildDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go mod tidy failed: %s\n%s", err, output)
	}

	outWasm := filepath.Join(buildDir, "script.wasm")
	absOutWasm, err := filepath.Abs(outWasm)
	if err != nil {
		absOutWasm = outWasm
	}

	var cmd *exec.Cmd
	if c.UseTinyGo && isTinyGoAvailable() {
		cmd = exec.Command("tinygo", "build", "-o", absOutWasm, "-target", "wasm", ".")
		cmd.Dir = buildDir
	} else {
		cmd = exec.Command("go", "build", "-o", absOutWasm, ".")
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

	if _, err := os.Stat(outWasm); os.IsNotExist(err) {
		entries, _ := os.ReadDir(buildDir)
		var files []string
		for _, e := range entries {
			files = append(files, e.Name())
		}
		return nil, fmt.Errorf("wasm file not generated at %s, build output: %s, files in buildDir: %v", outWasm, output, files)
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
	exePath, err := os.Executable()
	if err == nil {
		exePath, _ = filepath.EvalSymlinks(exePath)
		dir := filepath.Dir(exePath)
		for {
			goModPath := filepath.Join(dir, "go.mod")
			if data, err := os.ReadFile(goModPath); err == nil {
				if strings.Contains(string(data), "module github.com/cameron-webmatter/galaxy") {
					return dir
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/cameron-webmatter/galaxy")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return ""
}

func prepareScript(script, hash string, useTinyGo bool) (string, error) {
	imports := extractImports(script)
	body := removeImports(script)
	body = removePackageDecl(body)
	hasMain := containsMainFunc(body)

	var final strings.Builder
	final.WriteString("package main\n\n")

	if len(imports) > 0 {
		final.WriteString("import (\n")
		for _, imp := range imports {
			final.WriteString(fmt.Sprintf("\t%s\n", imp))
		}
		final.WriteString(")\n\n")
	}

	if !hasMain {
		vars, funcs, execCode := separateFunctionsFromCode(body)

		for _, v := range vars {
			final.WriteString(v)
			final.WriteString("\n")
		}
		if len(vars) > 0 {
			final.WriteString("\n")
		}

		for _, fn := range funcs {
			final.WriteString(fn)
			final.WriteString("\n\n")
		}

		final.WriteString("func main() {\n")
		if execCode != "" {
			final.WriteString(indentCode(execCode))
			final.WriteString("\n")
		}
		final.WriteString("\tselect {}\n")
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

func separateFunctionsFromCode(body string) (variables []string, functions []string, executableCode string) {
	lines := strings.Split(body, "\n")
	var funcLines []string
	var execLines []string
	inFunc := false
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFunc && strings.HasPrefix(trimmed, "func ") {
			inFunc = true
			funcLines = append(funcLines, line)
			braceDepth = strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth == 0 {
				inFunc = false
			}
			continue
		}

		if inFunc {
			funcLines = append(funcLines, line)
			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")
			if braceDepth == 0 {
				inFunc = false
				functions = append(functions, strings.Join(funcLines, "\n"))
				funcLines = nil
			}
			continue
		}

		if trimmed != "" {
			if strings.HasPrefix(trimmed, "var ") || strings.HasPrefix(trimmed, "const ") {
				variables = append(variables, line)
			} else {
				execLines = append(execLines, line)
			}
		}
	}

	executableCode = strings.Join(execLines, "\n")
	return variables, functions, executableCode
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
