# Gastro

Go-powered web framework inspired by AstroJS. Combines Go's runtime performance with Astro's elegant component model.

## Quick Start

```bash
go run cmd/gastro/main.go test examples/basic/pages/index.gxc
go run cmd/gastro/main.go dev examples/basic/pages
```

## Project Structure

```
gastro/
├── cmd/gastro/          # CLI tool
├── pkg/
│   ├── parser/          # .gxc file parser
│   ├── executor/        # Go frontmatter executor
│   ├── template/        # Template rendering engine
│   └── router/          # File-based routing
└── examples/basic/      # Example project
```

## Component Format (.gxc)

```gastro
---
var title = "Hello"
var items = []string{"A", "B", "C"}
---
<h1>{title}</h1>
<ul>
  <li gastro:for={item in items}>{item}</li>
</ul>

<style scoped>
h1 { color: blue; }
</style>
```

## Features Implemented (Phase 1)

- ✅ .gxc file parsing
- ✅ Go frontmatter execution (variables, arithmetic, arrays)
- ✅ Template expressions `{variable}`
- ✅ Control flow directives (`gastro:if`, `gastro:for`)
- ✅ Slot system (default & named)
- ✅ File-based routing (static, dynamic `[id]`, catch-all `[...slug]`)
- ✅ Route priority & matching
- ✅ Basic CLI commands

## Roadmap

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for full 10-phase plan.
