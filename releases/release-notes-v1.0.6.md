# PPHLX Compiler v1.0.6 — Multi-Target Outputs & CLI Overrides

We are pleased to announce the release of **PPHLX v1.0.6**. This version introduces high-performance multi-target compilation options, on-the-fly CLI target overrides, cross-compilation configurations for Go standalone binaries, a new default dev server port, and refined development server execution bypasses.

## What's New in v1.0.6

### 🚀 Multi-Target Output Generation
Starting with v1.0.6, PPHLX can compile your component-based templates into multiple structural formats beyond standard PHP:
*   **Standard PHP (`php`)**: The default component hydration view engine.
*   **Standalone Go Binary (`standalone`)**: Compiles your entire application (pages, templates, assets, and routes) directly into a single self-contained executable binary file (`app` or `app.exe`).
*   **Static Site Generator (`ssg`)**: Generates fully-rendered static HTML, JS, and CSS pages, resolving PHP scripting blocks at build time using the local PHP CLI interpreter.
*   **Framework views (`blade` and `twig`)**: Automatically translates `.pphx` components into framework-native syntax (`<x-...>` for Laravel Blade and `{% include ... %}` for Twig).

### ⚡ On-The-Fly CLI Overrides
You can now quickly switch compilation target formats without modifying your configuration files by using the `--target` or `-t` flag:
```bash
# Compile to a single standalone binary
pphlx build --target standalone

# Compile directly to static HTML
pphlx build --target ssg
```

### 🧩 Environment Cross-Compilation
Configure cross-compilation target environments directly inside `pphlx.config.json` via the `"output"` block:
```json
{
  "srcDir": "src",
  "outDir": "dist",
  "output": {
    "target": "standalone",
    "goos": "linux",
    "goarch": "amd64"
  }
}
```
This enables compiling standalone binaries for foreign environments (e.g. Linux servers) directly from your local development machine (e.g. Windows).

### ⚡ High-Performance Dev Server UX
*   **Dev Mode Rebuild Bypass**: During `pphlx dev`, the Go standalone compiler bypasses active `go build` routines, eliminating binary build wait times and keeping local template watch-mode updates sub-second.
*   **Brand Port `6321`**: Custom dev server port with built-in TCP listener conflict detection. If `6321` is occupied, it automatically increments and tests `6322`, `6323`, etc.
*   **Console Interface**: Astro-style terminal layout with beautiful ANSI coloring.

---

## Supported Architectures (Included Assets)
This release includes precompiled native binaries and archives for the following architectures:
*   `pphlx-darwin-amd64.tar.gz` (macOS Intel)
*   `pphlx-darwin-arm64.tar.gz` (macOS Apple Silicon M1/M2/M3/M4/M5+)
*   `pphlx-linux-amd64.tar.gz` (Linux Intel)
*   `pphlx-linux-arm64.tar.gz` (Linux ARM)

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

### Start Development Server
```bash
npm run dev
```
