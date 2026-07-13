# PPHLX v1.0.0 — Initial Release

We are excited to announce the initial release of **PPHLX v1.0.0**, a high-performance web component compiler that transforms modern component-based layouts into standalone, production-ready PHP pages. 

PPHLX is designed to bring modern frontend workflows (components, layout templates, hot reloading) to standard PHP environments with **zero runtime overhead** in production.

---

## Key Features

*   **🚀 High-Performance Go Engine**: The core compiler is written in Go, compiling templates and script islands in milliseconds.
*   **🧩 Zero-Dependency PHP Monolith Output**: Generates clean, standalone `.php` files. No Node.js runtime, `node_modules`, or complex dev-server requirements on your production web host.
*   **⚡ WebAssembly (WASI) Supported**: Shipped with a built-in WASI loader, enabling the compiler to run seamlessly in any Node.js environment via standard tools.
*   **🔥 Hot-Reloading Dev Server**: Includes a local watch server for instant page hydration updates during development.
*   **🛠️ First-Class Tooling Ecosystem**:
    *   **NPM Package**: Standard package manager support (`npm install pphlx`).
    *   **VS Code Extension**: Dedicated syntax highlighting, brace-pipe scope parsing, and autocomplete for `.pphx` templates.
    *   **Homebrew Tap**: Fast macOS and Linux installations via `brew install pphlx`.

---

## Supported Architectures (Included Binaries)

This release includes precompiled native binaries and archives for the following architectures:
*   `pphlx-darwin-amd64.tar.gz` (macOS Intel)
*   `pphlx-darwin-arm64.tar.gz` (macOS Apple Silicon M1/M2/M3)
*   `pphlx-linux-amd64.tar.gz` (Linux Intel)
*   `pphlx-linux-arm64.tar.gz` (Linux ARM)

---

## Quick Start

### Installation (macOS & Linux via Homebrew)
```bash
brew tap pphlx/tap
brew install pphlx
```

### Installation (Node.js / NPM)
```bash
npm install pphlx
```

### Compile Project
```bash
pphlx build
```
