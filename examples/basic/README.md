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
- Template directives (`galaxy:if`, `galaxy:for`)
- Go frontmatter execution

## Run

```bash
# Dev server
galaxy dev

# Build static site
galaxy build
```

## Routes

- `/` - Homepage
- `/about` - About page
- `/components-test` - Component composition demo
