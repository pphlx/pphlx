# PPHLX Language Service (VS Code Extension)

> Rich editing experience for PPHLX templates, bringing syntax highlighting, brackets closing, and language support.

This is the official Visual Studio Code extension providing native language support and syntax configurations for **PPHLX (`.pphx`)** templates — the component-based standalone web compiler for PHP (similar to Astro for Node.js).

---

## Features

This extension provides a rich editing experience for PPHLX templates, including:

### 🎨 Syntax Highlighting
*   **Brace-Pipe Scopes**: Full context-aware highlighting for logical statements `{| ... |}` and echo values `{|= ... |}`.
*   **Embedded Languages**: Seamless color-coding token injection for HTML5, PHP, CSS, and client-side JavaScript.
*   **Directive Recognition**: Highlights custom imports like `@import Layout from '...'` natively.

### ✍️ Intelligent Editing & Auto-Closing Pairs
*   **Brace-Pipe Auto-Close**: Type `{|` or `{|=` and the editor will automatically insert the matching closing delimiter ` |}` and center your cursor.
*   **Workspace Integration**: Automatically maps any `.pphx` files in your workspace with the PPHLX language identifier and custom file icons.

---

## Installation

### From VSIX Package (Manual Build)
Download the `.vsix` file for the release that you want to install from the [GitHub Releases tab](https://github.com/pphlx/vscode/releases).

1. Open Visual Studio Code.
2. Go to the Extensions tab (`Ctrl+Shift+X` / `Cmd+Shift+X`).
3. Click the `...` menu in the upper right corner of the Extensions tab.
4. Select **Install from VSIX...** and select the `.vsix` file you just downloaded.

The extension can also be installed with the following command line:
```bash
code --install-extension /path/to/pphlx-1.0.4.vsix
```

---

## Configuring Workspace Options
To ensure auto-complete and bracket closing work flawlessly, make sure your editor's auto-closing brackets setting is enabled. In your `.vscode/settings.json`, add:

```json
{
  "editor.autoClosingBrackets": "always"
}
```

### Manual Language Association
If a `.pphx` file does not open with PPHLX highlighting automatically:
1. Click the Language Selector in the bottom-right status bar.
2. Select **Configure File Association for '.pphx'...**
3. Choose **PPHLX** from the list.

---

## Versioning
This language service extension is version-matched with the core PPHLX compiler suite. It is recommended to use matching versions of the composer package and the vscode extension to prevent any syntax mismatches.
