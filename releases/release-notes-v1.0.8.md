# PPHLX Compiler v1.0.8 — Parallel Environments Engine & Multi-Target Scaffolding

We are pleased to announce the release of **PPHLX v1.0.8**. This version introduces a native Go-powered parallel multi-target compilation engine, custom configuration profiles, native CLI flag overrides, and full-stack Windows/macOS/Linux/Android/iOS release orchestration.

## What's New in v1.0.8

### ⚡ Parallel Multi-Environment Compiling (`--all`)
*   **Zero-Dependency Concurrency**: Define a root-level `"environments"` map inside `pphlx.config.json` to manage profiles for your web server, native desktop application, and mobile workspaces in one unified file.
*   **Go Goroutine Parallelism**: Spawn light Go Goroutines on multiple CPU threads to build all environment targets concurrently:
    ```bash
    pphlx build --all
    ```
    Compiling all platforms concurrently runs at Go-native speeds, taking practically the same sub-second time as a single-target compile.
*   **Specific Profile Builds**: Build a single named configuration environment easily:
    ```bash
    pphlx build --env desktop-win
    ```

### 💻 Native Desktop Target (`desktop`)
Package your web codebases into a single, installable native desktop application window with an embedded micro web server:
*   **Edge WebView2 Integration on Windows**: Uses `go-webview2` for pure-Go win32 bindings, enabling **CGO-free compilation on Windows** out-of-the-box, resulting in extremely fast execution and compact binary sizes (~9.4MB).
*   **WebKit Integration on macOS & Linux**: Integrates with native Safari WebKit (Cocoa) or GTK WebKit via `webview_go`.

### 🔌 Standard Desktop Drivers & Custom Go Bridges
*   **Built-in Drivers**: Exposes the `pphlx.desktop` JS object in the browser runtime, allowing layouts to trigger native OS file pickers (`openFileDialog`, `saveFileDialog`), system notifications (`showNotification`), and window termination (`close`).
*   **Custom Extensions**: Automatically scans the project's `src/desktop/` directory for any custom Go extensions (`*.go`) and compiles them dynamically, automatically registering and exposing their structs into the frontend JavaScript context via a unified `DesktopWindow` interface.

### 📱 Native Mobile Target Scaffolding (`android` & `ios`)
Scaffold complete native workspaces directly from the PPHLX CLI:
*   **Android Target (`android`)**: Natively scaffolds a complete Gradle Android Studio project layout in `dist/android/` pre-loaded with Java WebView configurations, Android Manifest files, and precompiled static HTML/JS/CSS assets.
*   **iOS Target (`ios`)**: Natively scaffolds an Xcode project structure inside `dist/ios/` with Swift view controllers, AppDelegate lifecycles, and preloaded asset bundles in `www/`.

---

## Supported Architectures (Included Assets)
This release includes precompiled native binaries and archives for the following architectures:
*   `pphlx-darwin-amd64.tar.gz` (macOS Intel)
*   `pphlx-darwin-arm64.tar.gz` (macOS Apple Silicon M1/M2/M3/M4/M5)
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

### Run Parallel Compilations
```bash
pphlx build --all
```
