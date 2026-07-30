# PPHLX Compiler Engine v1.1.6 Release Notes

PPHLX v1.1.6 introduces **Single Source of Truth** JavaScript configuration (`pphlx.config.mjs` / `pphlx.config.cjs`) with Astro parity, zero-config default fallbacks, dynamic framework plugin auto-detection from `package.json`, root-absolute SSG asset path traversal invariants, and non-blocking static asset diagnostics.

---

### Key Upgrades & Fixes in v1.1.6

- **Single Source of Truth Configuration (`pphlx.config.mjs` / `pphlx.config.cjs`)**: Consolidates compiler settings and Vite island plugins into a single JavaScript/TypeScript file without requiring separate `.vite.config.mjs` files.
- **Astro Parity Zero-Config Mode**: Automatically applies default settings (`srcDir: "src"`, `outDir: "dist"`, `base: "/"`, `target: "php"`) when fields are omitted or when running without a configuration file (`export default defineConfig({})`).
- **Dynamic Framework Plugin Auto-Detection (`detectVitePlugins`)**: Inspects `package.json` dependencies dynamically to import only installed UI framework plugins (`vue`, `svelte`, `solid-js`, `react`, `preact`).
- **SSG Deep Route Asset Path Traversal Invariant**: Injected `<link rel="stylesheet">` and `<script src="...">` tags now use root-absolute (`/assets/...`) or `base`-prefixed paths, resolving 404 resource errors on deep subpages (e.g. `/pages/gradient/index.html`).
- **Non-Blocking Public & Static Asset Resolution**: Replaced fatal build halts (`[FATAL BUILD HALT]` / `os.Exit(1)`) with non-blocking diagnostic warnings for missing or dynamic static asset paths.
- **Clean Root Directory Isolation**: Stores temporary auto-generated Vite bundler configuration files inside `.pphlx/pphlx.vite.config.mjs`, keeping project root directories 100% clean.

---

### Release Binary SHA256 Checksums

| Target Archive | SHA256 Checksum |
| :--- | :--- |
| `pphlx-darwin-arm64.tar.gz` | `3a5e5fd1d55e5df5d39d12f6570ec24d8f9342a2bdaaded4f563b008d3082bc0` |
| `pphlx-darwin-amd64.tar.gz` | `a87d0a825ba1e6e60f54a8c6948c6cb51f9d70b49ff500faeef9ad76287cf176` |
| `pphlx-linux-arm64.tar.gz` | `8b23cb7d17ece488c755dd49df0f1b4f579cb01d35e9872b524e308aa2bbeb7e` |
| `pphlx-linux-amd64.tar.gz` | `367eb33a0ac62a44827122e2768f561d5aa50e20f59a168fcdbd188cceae30c4` |
| `pphlx-windows-amd64.zip` | `5b49ecf5fb8a4d20e6b68561caa571dd426a144affb694d2c82f2f06bfcfed3c` |
