# Gastro 🚀

A blazing-fast, Go-powered web framework inspired by Astro. Build content-focused websites with ease.

## Features

- **🚀 Fast** - Built with Go for maximum performance
- **📦 Zero Config** - Sensible defaults, optional configuration
- **🔥 Hot Reload** - Instant updates during development
- **🎨 Component-Based** - Reusable `.gxc` components
- **⚡ Static Site Generation** - Pre-render for lightning speed
- **🛠️ Rich CLI** - Powerful command-line interface
- **📱 Interactive Setup** - Guided project creation

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/gastro/gastro
cd gastro
go build -o gastro ./cmd/gastro

# Add to PATH (optional)
sudo mv gastro /usr/local/bin/
```

### Create a New Project

```bash
gastro create my-project
cd my-project
gastro dev
```

Visit `http://localhost:4322` 🎉

## Commands

### `gastro create [name]`
Create a new project with interactive setup.

```bash
gastro create my-site
```

**Templates:** minimal, blog, portfolio, documentation

### `gastro dev`
Start development server with hot reload.

```bash
gastro dev                    # Start on port 4322
gastro dev --port 3000        # Custom port
gastro dev --open             # Auto-open browser
```

**Hotkeys:**
- `o + enter` - Open browser
- `r + enter` - Restart server
- `c + enter` - Clear console
- `q + enter` - Quit

### `gastro build`
Build for production.

```bash
gastro build                  # Output to ./dist
gastro build --outDir ./out   # Custom output
gastro build --verbose        # Show details
```

### `gastro preview`
Preview production build locally.

```bash
gastro preview                # Serve on port 4323
gastro preview --port 8080    # Custom port
gastro preview --open         # Auto-open browser
```

### `gastro add [integration]`
Add integrations to your project.

```bash
gastro add tailwind           # Add Tailwind CSS
gastro add                    # Interactive selection
```

**Available:** react, vue, svelte, tailwind, sitemap

### `gastro check`
Validate your project for errors.

```bash
gastro check                  # Check all .gxc files
gastro check --verbose        # Show details
```

### `gastro info`
Display environment information.

```bash
gastro info
```

### `gastro sync`
Sync types and configuration.

```bash
gastro sync
```

### `gastro docs`
Open documentation in browser.

```bash
gastro docs
```

## Project Structure

```
my-project/
├── pages/              # Routes (file-based routing)
│   ├── index.gxc       # / route
│   └── about.gxc       # /about route
├── components/         # Reusable components
│   └── Layout.gxc
├── public/            # Static assets
│   └── style.css
├── gastro.config.json # Configuration
└── package.json       # NPM dependencies
```

## Component Syntax (.gxc)

```gxc
---
title: string = "My Page"
---

<Layout title={title}>
  <main>
    <h1>Welcome to Gastro!</h1>
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
  console.log('Hello from Gastro!');
</script>
```

## Configuration

`gastro.config.json`:

```json
{
  "site": "https://example.com",
  "base": "/",
  "outDir": "./dist",
  "server": {
    "port": 4322,
    "host": "localhost"
  }
}
```

## Global Flags

All commands support:
- `--root <path>` - Project root directory
- `--config <path>` - Config file path
- `--verbose` - Verbose logging
- `--silent` - Silent mode
- `--help` - Show help
- `--version` - Show version

## Examples

See `examples/` directory:
- `examples/basic` - Simple site with multiple pages
- `examples/ssr` - Server-side rendering demo

## Development

```bash
# Build from source
go build -o gastro ./cmd/gastro

# Run tests
go test ./...

# Build example
cd examples/basic
gastro build
```

## Features

### File-Based Routing
- `pages/index.gxc` → `/`
- `pages/about.gxc` → `/about`
- `pages/blog/[slug].gxc` → `/blog/:slug` (dynamic)

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

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT License - see [LICENSE](LICENSE)

## Credits

Inspired by [Astro](https://astro.build) - Built with Go
