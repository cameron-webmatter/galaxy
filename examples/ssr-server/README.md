# SSR Server Example

Full server-side rendering with on-demand page generation.

## Features

- Server-mode output (`type = "server"`)
- Pages rendered on each request
- Access to Request context
- Middleware support (coming soon)
- API endpoints (coming soon)

## Run

```bash
# Dev mode
galaxy dev

# Build server binary
galaxy build

# Run built server
./dist/server/server
```

## Output

Builds a standalone Go binary that serves pages on-demand.
