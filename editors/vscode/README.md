# GXC Language Support for VS Code

Syntax highlighting and language server support for `.gxc` files (Galaxy components).

## Features

- **Syntax Highlighting**: Multi-layered highlighting for Go frontmatter, HTML, CSS, and JavaScript
- **Diagnostics**: Real-time error checking for Go syntax and undefined variables
- **Auto-completion**: Variables from frontmatter, galaxy directives
- **Hover Info**: Type information and values for variables

## Installation

### From Source

1. Build LSP server:
```bash
cd /path/to/galaxy
go build -o galaxy ./cmd/galaxy
```

2. Install extension dependencies:
```bash
cd editors/vscode
npm install
npm run compile
```

3. Link extension (development):
```bash
ln -s $(pwd) ~/.vscode/extensions/gxc-language-0.1.0
```

OR package and install:
```bash
npm install -g vsce
vsce package
code --install-extension gxc-language-0.1.0.vsix
```

## Configuration

Set path to `galaxy` binary in VS Code settings:

```json
{
  "gxc.lsp.enable": true,
  "gxc.lsp.serverPath": "/path/to/galaxy"
}
```

## Syntax Regions

- **Frontmatter** (Go code between `---`): Full Go syntax highlighting
- **Template** (HTML): Standard HTML + custom directives
- **Expressions** (`{variable}`): Go expression highlighting
- **Directives** (`galaxy:if`, `galaxy:for`): Custom attributes
- **Scripts** (`<script>`): JavaScript highlighting
- **Styles** (`<style>`): CSS highlighting

## Example

```gxc
---
var title = "Hello World"
var items = []string{"A", "B", "C"}
---
<h1>{title}</h1>
<ul>
  <li galaxy:for={item in items}>{item}</li>
</ul>

<style scoped>
h1 { color: blue; }
</style>
```

## Development

Watch TypeScript changes:
```bash
npm run watch
```

Reload window in VS Code: `Cmd+R` (Mac) or `Ctrl+R` (Windows/Linux)
