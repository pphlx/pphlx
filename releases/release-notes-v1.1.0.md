# PPHLX v1.1.0 Release Notes

PPHLX `v1.1.0` introduces multi-stack dev server port conflict auto-increment with 150ms PHP process health checks, root-level scaffolding support (`layouts/`, `components/`, `index.pphx`), dynamic Vite island entry generation for monolithic structures, and recursive file watcher filtering.

## 🚀 Key Highlights & Enhancements

### 1. Multi-Stack Port Auto-Increment & Process Health Monitor
Automatic socket availability checking across IPv4 (`127.0.0.1`), IPv6 (`[::1]`), and wildcard (`0.0.0.0`) interfaces. Added a 150ms `php.exe` process health check in `startDevServerAndWatcher()`. If a port bind collision occurs, `pphlx dev` auto-increments ports (`6321` → `6322` → `6323`) seamlessly, guaranteeing active dev servers across multiple simultaneous projects.

### 2. Root-Level Scaffolding (`pphlx init`)
Updated CLI scaffolding behavior to place `layouts/Layout.pphx`, `components/`, and `index.pphx` out of `src/` at the root project directory level for monolithic PHP structures without creating redundant `src/` folders.

### 3. Dynamic Vite Island Entry Path Resolution
Updated `runViteBuild` in the compiler engine to dynamically generate `.pphlx_entry.js` and output bundles relative to configured `srcDir` and `outDir`, seamlessly building Vue, Svelte, and SolidJS component islands when `"srcDir": "."`.

### 4. Recursive File Watcher Filtering
Fixed watcher logic to ignore `dist/`, `node_modules/`, `.git/`, `.vscode/`, `storage/`, and `.antigravity/`, preventing infinite rebuild loops in root-scaffolded projects.

---

## 📦 Binary Release Checksums (SHA256)

| Target Platform | Asset Filename | SHA256 Checksum |
| :--- | :--- | :--- |
| **macOS (Apple Silicon)** | `pphlx-darwin-arm64.tar.gz` | `66FE53524F0A3F3F5C5551C0E059A74AD0879192C63C969A224AD9259218C1AF` |
| **macOS (Intel x64)** | `pphlx-darwin-amd64.tar.gz` | `13173A68A36C8EF379398649F60B9335898F0F04F342F6CB07A7C24B0D6279C5` |
| **Linux (ARM64)** | `pphlx-linux-arm64.tar.gz` | `58956587A0C978C6FBF44C27F3A11377D9CBBA88731CA84926524EC872BC92A6` |
| **Linux (AMD64)** | `pphlx-linux-amd64.tar.gz` | `D2886F718250CA6F63223C7A59AB13F582913337C35DFE67A52D14F1C5CE390B` |
| **Windows (AMD64)** | `pphlx-windows-amd64.tar.gz` | `55CEE9217E7BC056B26F10AB2786093DFB5A44C48204F1EE03DD1E08CF10983C` |
