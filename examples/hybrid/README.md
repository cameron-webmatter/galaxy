# Hybrid Mode Example

Mix static and dynamic pages in one project.

## Features

- **Static pages** - Pre-rendered at build time
- **Dynamic pages** - Server-side rendered on-demand
- **Single deployment** - One binary serves both

## Pages

- `/` - Static (pre-rendered HTML)
- `/dynamic` - SSR (rendered per request)

## Configuration

```toml
[output]
type = "hybrid"
```

## Opt-out of Prerendering

Add to frontmatter:

```gxc
---
var title = "My Page"
// prerender = false
---
```

## Build & Run

```bash
# Build
galaxy build

# Static pages → dist/*.html
# Dynamic pages → dist/server/server binary

# Run server
./dist/server/server
```

Server serves static files first, falls back to SSR for dynamic routes.
