# PPHLX Compiler v1.0.7 — Native Desktop & Mobile Scaffolding Targets

We are pleased to announce the release of **PPHLX v1.0.7**. This version introduces native desktop application compilation capabilities, a CGO-free Webview engine for Windows, cross-platform standard driver integrations, and full Android/iOS native mobile scaffolding pipelines.

## What's New in v1.0.7

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

### Run Native Desktop Compile
```bash
pphlx build --target desktop
```
