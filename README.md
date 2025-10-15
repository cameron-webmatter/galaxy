# Galaxy 🚀

A blazing-fast, Go-powered web framework inspired by Astro. Build content-focused websites with ease.

## Features

- **🚀 Fast** - Built with Go for maximum performance
- **📦 Zero Config** - Sensible defaults, optional configuration
- **🔥 Hot Reload** - Instant updates during development
- **🎨 Component-Based** - Reusable `.gxc` components
- **⚡ Three Build Modes** - Static, Server (SSR), or Hybrid
- **🔌 Go-Powered Runtime** - Middleware & API endpoints in Go
- **🛠️ Rich CLI** - Powerful command-line interface
- **📱 Interactive Setup** - Guided project creation

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/galaxy/galaxy
cd galaxy
go build -o galaxy ./cmd/galaxy

# Add to PATH (optional)
sudo mv galaxy /usr/local/bin/
```

### Create a New Project

```bash
galaxy create my-project
cd my-project
galaxy dev
```

Visit `http://localhost:4322` 🎉

## Commands

### `galaxy create [name]`
Create a new project with interactive setup.

```bash
galaxy create my-site
```

**Templates:** minimal, blog, portfolio, documentation

### `galaxy dev`
Start development server with hot reload.

```bash
galaxy dev                    # Start on port 4322
galaxy dev --port 3000        # Custom port
galaxy dev --open             # Auto-open browser
```

**Hotkeys:**
- `o + enter` - Open browser
- `r + enter` - Restart server
- `c + enter` - Clear console
- `q + enter` - Quit

### `galaxy build`
Build for production.

```bash
galaxy build                  # Uses config output type
galaxy build --outDir ./out   # Custom output
galaxy build --verbose        # Show details
```

**Output depends on mode:**
- **Static:** HTML files in `./dist/`
- **Server:** Binary at `./dist/server/server`
- **Hybrid:** Static HTML + binary for dynamic routes

### `galaxy preview`
Preview production build locally.

```bash
galaxy preview                # Serve on port 4323
galaxy preview --port 8080    # Custom port
galaxy preview --open         # Auto-open browser
```

### `galaxy add [integration]`
Add integrations to your project.

```bash
galaxy add tailwind           # Add Tailwind CSS
galaxy add                    # Interactive selection
```

**Available:** react, vue, svelte, tailwind, sitemap

### `galaxy check`
Validate your project for errors.

```bash
galaxy check                  # Check all .gxc files
galaxy check --verbose        # Show details
```

### `galaxy info`
Display environment information.

```bash
galaxy info
```

### `galaxy sync`
Sync types and configuration.

```bash
galaxy sync
```

### `galaxy docs`
Open documentation in browser.

```bash
galaxy docs
```

## Project Structure

```
my-project/
├── src/
│   ├── pages/          # Routes (file-based routing)
│   │   ├── index.gxc   # / route
│   │   ├── about.gxc   # /about route
│   │   └── api/        # API endpoints (server/hybrid)
│   │       └── hello.go # /api/hello endpoint
│   ├── components/     # Reusable components
│   │   └── Layout.gxc
│   └── middleware.go   # Middleware (server/hybrid)
├── public/             # Static assets
│   └── style.css
├── galaxy.config.toml  # Configuration
└── go.mod              # Go dependencies (server/hybrid)
```

## Component Syntax (.gxc)

```gxc
---
title: string = "My Page"
---

<Layout title={title}>
  <main>
    <h1>Welcome to Galaxy!</h1>
    <p>Fast, Go-powered web framework</p>
  </main>
</Layout>

<style>
  main {
    max-width: 800px;
    margin: 0 auto;
  }
</style>

<script>
  console.log('Hello from Galaxy!');
</script>
```

## Build Modes

Galaxy supports three output modes configured in `galaxy.config.toml`:

### Static (SSG)
Pre-render all pages at build time. No server required.

```toml
[output]
type = "static"  # Default
```

**Output:** HTML files in `./dist/`

### Server (SSR)
Render pages on-demand with full server-side capabilities.

```toml
[output]
type = "server"

[adapter]
name = "standalone"  # or "cloudflare", "netlify", "vercel"
```

**Output:** Single Go binary in `./dist/server/server`

**Features:**
- Request context in pages
- Go-based middleware (`src/middleware.go`)
- Go-based API endpoints (`pages/api/*.go`)

### Hybrid (SSG + SSR)
Mix static and dynamic pages in one project.

```toml
[output]
type = "hybrid"

[adapter]
name = "standalone"
```

**By default:** All pages pre-rendered  
**Opt-out:** Add `// prerender = false` to frontmatter for SSR

## Configuration

`galaxy.config.toml`:

```toml
site = ""
base = "/"
outDir = "./dist"

[output]
type = "static"  # "static", "server", or "hybrid"

[server]
port = 4322
host = "localhost"

[adapter]
name = "standalone"  # For server/hybrid modes

[[plugins]]
name = "tailwindcss"
```

## Plugins

Galaxy supports an Astro-style plugin system for extending functionality.

### Available Plugins

- **tailwindcss** - Tailwind CSS integration with automatic processing
- **react** - React component islands (coming soon)
- **vue** - Vue component islands (coming soon)
- **svelte** - Svelte component islands (coming soon)

### Using Plugins

Add plugins to `galaxy.config.toml`:

```toml
[[plugins]]
name = "tailwindcss"
```

Or use the CLI:

```bash
galaxy add tailwind
```

### Tailwind CSS Plugin

Automatically processes `@tailwind` directives during build.

**Setup:**
1. Run `galaxy add tailwind` or manually install
2. Add plugin to config
3. Add directives to your CSS:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

The plugin will automatically process these during build.

## Global Flags

All commands support:
- `--root <path>` - Project root directory
- `--config <path>` - Config file path
- `--verbose` - Verbose logging
- `--silent` - Silent mode
- `--help` - Show help
- `--version` - Show version

## Middleware (Server/Hybrid Mode)

Create `src/middleware.go` to add middleware:

**Single middleware:**
```go
package src

import (
    "time"
    "github.com/galaxy/galaxy/pkg/middleware"
)

func OnRequest(ctx *middleware.Context, next func() error) error {
    ctx.Set("timestamp", time.Now().Format(time.RFC3339))
    return next()
}
```

**Multiple middleware (chained):**
```go
package middleware

import "github.com/galaxy/galaxy/pkg/middleware"

func LoggingMiddleware(ctx *middleware.Context, next func() error) error {
    // logging
    return next()
}

func AuthMiddleware(ctx *middleware.Context, next func() error) error {
    // auth
    return next()
}

// Chain multiple middleware with Sequence
func Sequence() []middleware.Middleware {
    return middleware.Sequence(
        LoggingMiddleware,
        AuthMiddleware,
    )
}
```

Access in `.gxc` pages:

```gxc
<p>Timestamp: {Locals.timestamp}</p>
<p>User: {Locals.user}</p>
```

## API Endpoints (Server/Hybrid Mode)

Create Go files in `src/pages/api/`:

```go
// src/pages/api/hello.go
package api

import "github.com/galaxy/galaxy/pkg/endpoints"

func GET(ctx *endpoints.Context) error {
    return ctx.JSON(200, map[string]string{
        "message": "Hello from Galaxy!",
    })
}

func POST(ctx *endpoints.Context) error {
    var body map[string]interface{}
    if err := ctx.BindJSON(&body); err != nil {
        return ctx.JSON(400, map[string]string{"error": "Invalid JSON"})
    }
    return ctx.JSON(200, body)
}
```

**Endpoints available at:** `/api/hello`

## Examples

See `examples/` directory:
- `examples/basic` - Static site with multiple pages
- `examples/ssr` - Full SSR with middleware & endpoints
- `examples/ssr-server` - Server-only mode demo
- `examples/hybrid` - Mixed static/dynamic pages

## Development

```bash
# Build from source
go build -o galaxy ./cmd/galaxy

# Run tests
go test ./...

# Build example
cd examples/basic
galaxy build
```

## Features

### File-Based Routing
- `src/pages/index.gxc` → `/`
- `src/pages/about.gxc` → `/about`
- `src/pages/blog/[slug].gxc` → `/blog/:slug` (dynamic)

### Components
- Reusable `.gxc` components
- Scoped styles with automatic hashing
- Props and frontmatter
- Layout components

### Assets
- Automatic CSS bundling
- Script bundling with hydration
- Static file serving from `public/`
- Asset optimization

### Hot Reload
- File watching for pages & components
- Instant browser updates
- Fast rebuilds

### Server-Side Rendering (SSR)
- On-demand page rendering
- Request context in templates
- Go-based middleware
- API endpoints in Go
- Single binary deployment
- No Node.js required

## Contributing

### Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/galaxy/galaxy
cd galaxy
```

2. **Install dependencies**
```bash
go mod download
```

3. **Install watchexec** (for hot reloading)
```bash
brew install watchexec  # macOS
# or
cargo install watchexec-cli  # Cross-platform
```

4. **Start development with hot reload**
```bash
# Terminal 1 - watches for changes and rebuilds
make watch

# Terminal 2 - test your changes
cd examples/basic
galaxy dev
```

The `make watch` command automatically rebuilds and installs the `galaxy` CLI to your `$GOBIN` whenever you modify any `.go` files in `pkg/`, `cmd/`, or `internal/`.

### Other Commands

```bash
make install    # Install galaxy CLI
make build      # Build binary to ./galaxy
make test       # Run tests
make clean      # Clean build artifacts
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details

## Editor Support

### NeoVim

Full support for `.gxc` files including syntax highlighting and LSP. See [editors/nvim/README.md](editors/nvim/README.md) for setup instructions.

**Quick Setup:**
```bash
# Copy plugin files
cp -r editors/nvim/* ~/.config/nvim/

# Add to init.lua
require'lspconfig'.gxc.setup{}
```

### VSCode

Install the GXC extension from `editors/vscode/`.

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Inspired by [Astro](https://astro.build) - Built with Go
