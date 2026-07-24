# PPHLX v1.1.1 Release Notes

PPHLX `v1.1.1` introduces full cross-platform native CLI execution on macOS (arm64 & amd64) and Linux/WSL (arm64 & amd64), WASI polyfill compatibility, cross-platform Vite shell execution delegation (`sh -c` / `cmd /c`), sub-millisecond build timing logs, and official support for 1-line project scaffolding via `create-pphlx` (`npm create pphlx@latest`).

## 🚀 Key Highlights & Enhancements

### 1. Cross-Platform Native CLI Execution (macOS & Linux/WSL)
Updated Node/PHP CLI runners to automatically resolve platform-native Go binaries (`pphlx-win.exe`, `pphlx-macos-arm64/amd64`, `pphlx-linux-arm64/amd64`), applying `chmod 755` executable permissions automatically on Unix systems before spawning.

### 2. Cross-Platform Vite Delegation (`sh -c` vs `cmd /c`)
Updated `main.go` compiler engine to select `cmd /c` on Windows and `sh -c` on Linux/macOS, resolving `exec: "cmd": executable file not found in $PATH` errors during Vite island compilation passes on non-Windows environments.

### 3. WASI Fallback Polyfill
Polyfilled `wasi_snapshot_preview1` via Node.js `wasi` module to eliminate WASM instantiation errors on macOS, Linux, and WSL environments when WASM fallback mode is active.

### 4. Sub-Millisecond & Full Pipeline Build Timing
Added precision build timing logs for native PHP template compilation (`✓ Built in 0.8ms`), Standalone Go binary compilation (`in 1.18s`), and total build completion (`in 3.5s`).

### 5. 1-Line Scaffolder Package (`create-pphlx`)
Released the official `create-pphlx` npm package (`npm create pphlx@latest`) for instant project initialization without relying on blocked `postinstall` scripts.

---

## 📦 Binary Release Checksums (SHA256)

| Target Platform | Asset Filename | SHA256 Checksum |
| :--- | :--- | :--- |
| **macOS (Apple Silicon)** | `pphlx-darwin-arm64.tar.gz` | `a056037d6e8b246ae184eb6ce0dc1869299e23326a7be489e472423677dcfcdf` |
| **macOS (Intel x64)** | `pphlx-darwin-amd64.tar.gz` | `10ea4769b8d83c65decbb0d38f953c0a9c2cb9c25d33238c68f351cc3d942cb3` |
| **Linux (ARM64)** | `pphlx-linux-arm64.tar.gz` | `341ee69cdb0e14a841706b513353276441fde35686c5ccb3a7b455fe93d0c6a1` |
| **Linux (AMD64)** | `pphlx-linux-amd64.tar.gz` | `3646211f18c570b5db1458588e97fdde665fb85d49e5d509c77da7a39afde359` |
| **Windows (AMD64)** | `pphlx-windows-amd64.zip` | `fc37d581f55bed4aa209a89ef6a4784aca9848b2a8008dac944dd987c9d92790` |
