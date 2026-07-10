# Piplex support for Visual Studio Code

🧑🚀 Not sure what Piplex is? See our core repository at [github.com/KillerTyzon/Piplex](https://github.com/KillerTyzon/Piplex)!

This is the official Visual Studio Code extension providing native language support, syntax highlighting, and auto-closing configurations for **Piplex (`.pphx`)** templates.

---

## What is Piplex?

Piplex is a high-performance web compiler and native toolchain written in Go. It compiles component-based layouts into standard PHP monolith pages. It enables developers to use modern frontend frameworks (React, Vue 3, Svelte 4, SolidJS, and Preact) side-by-side on a single page, fully hydrated by server-side PHP data, with **zero production runtime dependencies**.

---

## Features

### 🎨 Native Syntax Highlighting
*   **Brace-Pipe Scopes**: Full context-aware highlighting for logical statements `{| ... |}` and echo values `{|= ... |}`.
*   **Embedded Languages**: Seamless color-coding token injection for HTML5, PHP, CSS, and client-side JavaScript.
*   **Directives Recognition**: Highlights custom imports like `@import Layout from '...'` natively.

### ✍️ Intelligent Editing & Auto-Closing Pairs
*   **Brace-Pipe Auto-Close**: Type `{|` or `{|=` and the editor will automatically insert the matching closing delimiter ` |}` and center your cursor.
*   **Automatic File Association**: Automatically maps any `.pphx` files in your workspace with the Piplex language identifier and custom file icons.

---

## Installation

### From VSIX Package (Manual Build)
1. Open Visual Studio Code.
2. Open the Extensions sidebar (`Ctrl+Shift+X` on Windows/Linux, `Cmd+Shift+X` on Mac).
3. Click the `...` (More Actions) menu in the top-right corner of the Extensions panel.
4. Click **Install from VSIX...** and select the `piplex-1.0.0.vsix` file.

### Via command line:
```bash
code --install-extension /path/to/piplex-1.0.0.vsix
```

---

## Configuration & Workspace Setup

To ensure auto-complete and bracket closing work flawlessly:

### Auto-Closing Settings
Make sure your editor's auto-closing brackets setting is enabled. In your `.vscode/settings.json`, add:
```json
{
  "editor.autoClosingBrackets": "always"
}
```

### Manual Language Association
If a `.pphx` file does not open with Piplex highlighting automatically:
1. Click the Language Selector in the bottom-right status bar.
2. Select **Configure File Association for '.pphx'...**
3. Choose **Piplex** from the list.

---

## Contributing & Development

We welcome community contributions to advance the developer experience of the Piplex extension. To test changes locally:
1. Clone the repository.
2. Install global packaging dependencies: `npm install -g @vscode/vsce`.
3. Package the extension: `vsce package`.
4. Install the newly generated `.vsix` file to verify updates.
