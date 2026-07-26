# PPHLX Compiler Engine v1.1.5 Release Notes

PPHLX v1.1.5 introduces formal language syntax specification compliance, robust component property JSON serialization for multi-framework islands, dev server PHP Stdin stream evaluation to eliminate Windows command line limits, and isolated `.pphlx/cache` dev server compilation mode.

---

### Highlights

- **PPHLX Language Specification Standard (v1.0)**: Established formal specification (`docs/SYNTAX_SPEC.md`) defining `.pphx` template structure, token semantics (`{|= $expr |}`, `{| $stmt |}`), island hydration directives (`client:load`, `client:visible`, `client:idle`), and asset injection invariants.
- **Component Prop JSON Serialization Invariant**: Attributes using `{|= $expr |}` (e.g. `<FeedbackCard title="{|= $reactTitle |}" client:load />`) are serialized into `json_encode($expr)` inside island script tags, guaranteeing 100% valid JSON payload evaluation for React, Vue, Svelte, SolidJS, and Preact islands.
- **Windows Command-Line 32KB Argument Truncation Fix**: Dev server streams raw HTML through `cmd.Stdin` (`php -r "eval('?>'.file_get_contents('php://stdin'));"`), eliminating Windows 32KB command line argument string limits.
- **`.pphlx/cache` Dev Mode Isolation**: Redirected dev mode Vite island compilation output to `.pphlx/cache/` (matching Astro's `.astro/cache` standard). Ensures production `dist/` is 100% untouched and never created during `npm run dev`.
- **Deterministic FNV-32a Island Hashing**: Replaced timestamp-based island IDs with deterministic FNV-32a component/framework hashing for 100% consistent island container IDs across page refreshes.
- **Single Source of Truth Go WASM Compiler Engine**: Refactored `main.go` and `main_wasm.go` so all page compilation and autonomous asset tag injection (`<script src="assets/js/bundle.js"></script>` and `<link rel="stylesheet" href="assets/css/styles.css">`) are executed via unified `CompilePageWithAssets()` helper in `main.go`.

---

### SHA256 Checksums

| Archive File | SHA256 Checksum |
| :--- | :--- |
| `pphlx-darwin-arm64.tar.gz` | `c167f09db07c1b2002e507712a344df512f76d39d82eb49d4801183b4febea75` |
| `pphlx-darwin-amd64.tar.gz` | `aeb52f4d4c5f2f73c15c6d1c8fffbcc8e8d1d9b3c7a32e329d5d581a0ba5d6b2` |
| `pphlx-linux-arm64.tar.gz` | `8b54fdf4f335a9296bd7f9fd5a73f2c6ad4d3e8ff3d3bf4848d220416c51323c` |
| `pphlx-linux-amd64.tar.gz` | `b93981dab184ee8050f13b190bd436b205d19c4e936750ced8102341d692de6d` |
| `pphlx-windows-amd64.zip` | `00845f0e29e41567122089f53e4dbbff1abf58faefc28bee432477ed96702081` |
