# Basic SSG Example

Static site generation (SSG) example.

## Structure

```
pages/        # Route files (.gxc)
components/   # Reusable components
public/       # Static assets
```

## Features

- File-based routing
- Component composition with auto-import
- Scoped CSS
- Template directives (`gastro:if`, `gastro:for`)
- Go frontmatter execution

## Run

```bash
# Dev server
gastro dev

# Build static site
gastro build
```

## Routes

- `/` - Homepage
- `/about` - About page
- `/components-test` - Component composition demo
