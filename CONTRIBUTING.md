# Contributing to Galaxy

Thank you for your interest in contributing to Galaxy! 🎉

## Getting Started

### Prerequisites
- Go 1.23 or later
- Git

### Setup
```bash
# Fork and clone the repo
git clone https://github.com/YOUR-USERNAME/galaxy
cd galaxy

# Build the project
go build -o galaxy ./cmd/galaxy

# Run tests
go test ./...
```

## Development Workflow

### 1. Create a Branch
```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-fix
```

### 2. Make Changes
- Write clean, readable code
- Follow Go conventions
- Add tests for new features
- Update documentation as needed

### 3. Test Your Changes
```bash
# Run all tests
go test ./...

# Test the CLI manually
./galaxy create test-project
cd test-project
../galaxy dev
```

### 4. Commit
```bash
git add .
git commit -m "feat: add new feature"
```

**Commit Message Format:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `chore:` Tooling/config changes

### 5. Push and Create PR
```bash
git push origin feature/my-feature
```

Then create a Pull Request on GitHub.

## Project Structure

```
galaxy/
├── cmd/galaxy/          # CLI entry point
├── pkg/
│   ├── cli/             # CLI commands
│   ├── server/          # Dev server
│   ├── build/           # Build system
│   ├── parser/          # .gxc parser
│   ├── compiler/        # Component compiler
│   ├── template/        # Template engine
│   ├── router/          # File-based router
│   ├── executor/        # Frontmatter executor
│   ├── ssr/             # SSR context
│   ├── templates/       # Project templates
│   └── prompts/         # Interactive prompts
├── internal/
│   └── assets/          # Asset bundler
├── examples/            # Example projects
└── templates/           # Embedded templates
```

## Areas to Contribute

### High Priority
- Framework integrations (React, Vue, Svelte islands)
- Watch mode for `check` command
- Type generation for `sync` command
- Content collections support
- API routes
- Middleware system

### Medium Priority
- More project templates
- Better error messages
- Performance optimizations
- Config file validation
- Documentation improvements

### Low Priority
- Telemetry system
- Preferences management
- Additional integrations
- Developer toolbar

## Coding Guidelines

### Go Style
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused

### CLI Commands
- Use Cobra framework
- Support global flags (`--verbose`, `--silent`, `--root`)
- Provide helpful error messages
- Add `--help` documentation

### Testing
- Write tests for new features
- Aim for high coverage on core packages
- Use table-driven tests where appropriate

### Documentation
- Update README.md for new features
- Add command documentation
- Include examples

## Code Review Process

1. Maintainer reviews PR
2. Feedback addressed
3. Tests must pass
4. Documentation updated
5. PR merged

## Questions?

- Open an issue for bugs
- Start a discussion for feature ideas
- Ask questions in issues/discussions

## License

By contributing, you agree your contributions will be licensed under the MIT License.

---

Thank you for contributing! 🚀
