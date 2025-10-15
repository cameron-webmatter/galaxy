# Galaxy 🚀

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
galaxy build                  # Output to ./dist
galaxy build --outDir ./out   # Custom output
galaxy build --verbose        # Show details
```

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
├── pages/              # Routes (file-based routing)
│   ├── index.gxc       # / route
│   └── about.gxc       # /about route
├── components/         # Reusable components
│   └── Layout.gxc
├── public/            # Static assets
│   └── style.css
├── galaxy.config.json # Configuration
└── package.json       # NPM dependencies
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

## Configuration

`galaxy.config.json`:

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
go build -o galaxy ./cmd/galaxy

# Run tests
go test ./...

# Build example
cd examples/basic
galaxy build
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
