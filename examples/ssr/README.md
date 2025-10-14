# SSR Example

Server-side rendering with request context and dynamic routes.

## Features

- **Request Context**: Access HTTP request data in templates
- **Dynamic Routes**: `[id]` and `[...slug]` patterns
- **Auto-import Components**: No explicit imports needed
- **Scoped CSS**: Component-level styling

## Run

```bash
# Dev server
gastro dev

# Static build (no Request context)
gastro build
```

## Routes

- `/` - Homepage with request info
- `/api/user/123` - Dynamic route with `[id]` param
- `/blog/2024/post` - Catch-all route with `[...slug]`
- `/components` - Component composition example

## Request Context

```gxc
<div gastro:if={Request}>
    <p>Path: {Request.Path()}</p>
    <p>Method: {Request.Method()}</p>
</div>
```

Available methods: `Path()`, `Method()`, `URL()`, `Param()`, `QueryParam()`, `Header()`
