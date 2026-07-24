# Changelog

All notable changes to the **PPHLX Compiler Core** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.1.1] - 2026-07-24

### Added
*   **Sub-Millisecond & Pipeline Build Timing**: Added precision build timing output (`✓ Built in 0.8ms`), Standalone Go binary compilation timing (`in 1.18s`), and total pipeline elapsed duration (`in 3.5s`).

### Fixed
*   **Cross-Platform Shell Execution for Vite Delegation**: Replaced hardcoded `cmd /c` with platform check (`runtime.GOOS == "windows"` vs `sh -c`), fixing Vite component bundling failures on Linux, macOS, and WSL environments (`exec: "cmd": executable file not found in $PATH`).
*   **Standalone Go Binary Cross-Platform Build Flags**: Ensured `GOOS` and `GOARCH` environment variable injection works smoothly across non-Windows operating systems during native compilation passes.

---

## [1.1.0] - 2026-07-23

### Added
*   **Clean Root Layout Scaffolding**: Default initialization (`pphlx init`) scaffolds `layouts/Layout.pphx`, `components/`, and `index.pphx` out of `src/` matching monolithic template architectures.

### Fixed
*   **Dev Server Port Allocation & Health Monitoring**: `isPortAvailable` checks IPv4 (`127.0.0.1`), IPv6 (`[::1]`), and wildcard (`0.0.0.0`) sockets. `pphlx dev` actively monitors `php.exe` health for 150ms after launch, automatically incrementing ports (`6321` -> `6322`) on binding collisions to prevent silent dev server failures.
*   **Diagnostic Configuration Errors**: Removed misleading `./test_project/pphlx.config.json` error fallback in favor of clear project configuration diagnostics.
*   **Recursive Watcher Optimization**: Folder watcher now filters `dist/`, `node_modules/`, and system dotfiles to eliminate infinite rebuild loops when `"srcDir": "."`.

---

## [1.0.9] - 2026-07-23

### Added
*   **Multi-Framework Islands Hydration Engine:** Native client hydration mounting support for **React, Preact, SolidJS, Vue 3, and Svelte 4** side-by-side on the same PHP page.
*   **Progressive Hydration:** Selective client hydration triggers (`client:load`, `client:visible`, `client:idle`).
*   **WebAssembly Playground Engine:** Real-time in-browser compilation support via `main_wasm.go` for live template editing at `https://pphlx.org/play`.
*   **Vite Compiler Delegation:** Automatic Vite build integration for compiling `.svelte`, `.vue`, `.jsx`, and `.tsx` islands during build phase.
*   **Brace-Pipe Delimiter Syntax:** `{|= [expression] |}` and `{| [statement] |}` collision-free template expressions.
*   **Asset Bundling:** Automatic extraction of inline `<style>` and `<script>` blocks from components into unified `app.css` and `app.js` output files.
*   **Multi-Target Compiler:** Native targets for `"php"`, `"standalone"` (Go binary server), `"desktop"` (WebView2/WebView), `"android"` (Gradle project scaffold), `"ios"` (Swift Xcode project scaffold), `"ssg"`, `"blade"`, and `"twig"`.
*   **Desktop Native OS Bridge:** Direct native filesystem (`openFileDialog`, `saveFileDialog`), system notification, and window control API (`pphlx.desktop.*`) with custom Go bridge extension support.
*   **SSE Streaming & State Bridge:** Server-Sent Events streaming bridge connecting backend PHP state with frontend client islands.
*   **Recursive Dev Watcher:** File change watcher with debouncing for sub-second template recompilation.
*   **Model Context Protocol (MCP) Server Integration:** Built-in MCP endpoints for documentation search, island generation, and best practice inspection.

---

## [1.0.0] - 2026-01-15

### Added
*   Initial open-source release of PPHLX Compiler Core.
