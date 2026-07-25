# PPHLX v1.1.4 Release Notes

PPHLX `v1.1.4` introduces a pure 100% in-memory Go HTTP dev server engine, 2-pass dependency graph component suppression, `.pphlxignore` git-style exclusion manifest, high-performance channel-buffered dev server HTTP request logging, flexible `srcDir` entry resolution, and formatted ready banner styling.

## 🚀 Key Highlights & Enhancements

### 1. Pure 100% In-Memory Go HTTP Dev Server Engine
Refactored `pphlx dev` to compile pages and evaluate templates 100% in-memory without creating `.pphlx_dev_cache/` or `.pphlx_router.php` files on disk. Development mode leaves `dist/` untouched with zero disk writes.

### 2. 2-Pass Dependency Graph Compilation & Component Suppression
Automatically tracks component dependencies via `@import` statements to inline template components into parent routes while suppressing duplicate component emissions in `dist/`.

### 3. Safe Empty `dist/` Directory Contents Wipe (`wipeDirContents`)
Safely clears files and subdirectories inside `dist/*` while preserving the root `dist/` directory handle for active dev servers and file explorers.

### 4. `.pphlxignore` Git-Style Exclusion Manifest
Support for `.pphlxignore` build exclusion rules with wildcard matching.

### 5. High-Performance Go-Optimized Dev Server Request Logger
Real-time HTTP request logging with channel-buffered non-blocking worker (`logChan`), zero-allocation object pooling (`sync.Pool`), sub-millisecond precision (`µs`/`ms`), and category badges (`(page)`, `(virtual)`, `(asset)`, `(missing)`).

### 6. Flexible `srcDir` Entry Resolution
Added support for configuring `srcDir` as either a directory (e.g., `"src"`, `"src/demo"`) or an explicit template file (e.g., `"src/index.pphx"`).

### 7. In-Memory Static Asset Fallback
Multi-tier fallback serving static assets from `public/` and `src/` directly in memory without disk copies, supporting large media streaming via HTTP Range requests.

---

## 📦 Binary Release Checksums (SHA256)

| Target Platform | Asset Filename | SHA256 Checksum |
| :--- | :--- | :--- |
| **macOS (Apple Silicon)** | `pphlx-darwin-arm64.tar.gz` | `2f3b304c0e3f685f51037cdba3b6ab56b02edc00f962f02356a8ab3e51d80b82` |
| **macOS (Intel x64)** | `pphlx-darwin-amd64.tar.gz` | `ca1319746b906d88bd1d97cf41d78796a49b8e575993abcb56cb397e9b30098a` |
| **Linux (ARM64)** | `pphlx-linux-arm64.tar.gz` | `f4a3e68d0b72cf513e2c49712347bf820420c67a304b88039bb16e02b1a41eeb` |
| **Linux (AMD64)** | `pphlx-linux-amd64.tar.gz` | `de2430c1bf194fff8c899451997dfcc8272aa040b18d927f12ff20e7289075a9` |
| **Windows (AMD64)** | `pphlx-windows-amd64.zip` | `fc5f9b9375f7c9cc8c08fa004c415499ef74c6342d0ff92dbdd586479a351524` |
