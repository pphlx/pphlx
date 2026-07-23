# PPHLX Product & Technical Roadmap

This document outlines the strategic vision and milestones for the **PPHLX Compiler Core** (`pphlx.org`), designed for commercial, production-grade applications, enterprise monoliths, and CI/CD pipelines.

---

## 🎯 Strategic Vision

PPHLX is a fast, zero-dependency multi-framework template compiler for PHP. It bridges modern frontend framework ergonomics (React, Vue, Svelte, Preact, SolidJS) with standard PHP backends with zero Node.js runtime overhead in production.

---

## 🗺️ Release Milestones

### Phase 1: Core Foundation & Multi-Framework Hydration — ✅ Completed
- [x] High-performance Go compiler engine & AST parser.
- [x] Collision-free Brace-Pipe syntax (`{|= ... |}` and `{| ... |}`).
- [x] Multi-Framework Islands Hydration (React, Preact, Vue 3, Svelte 4, SolidJS).
- [x] Vite delegation pass & automatic asset extraction (`app.css`, `app.js`).
- [x] Multi-target compilation (`php`, `standalone`, `desktop`, `android`, `ios`, `ssg`, `blade`, `twig`).
- [x] Native OS WebView drivers (`pphlx.desktop.*`) and WebAssembly browser compilation (`main_wasm.go`).

---

### Phase 2: Native AST Diagnostic Linter, Docker & CI/CD — 🎯 Active Development
- [ ] **Astro-Grade Native AST Diagnostic Engine**: Integrated compile-time linter checking image/asset resolution (`<img src="...">`), unclosed component tags, and Brace-Pipe syntax errors with precise line/column pointers.
- [ ] **Dev Watcher Zero-Write Halt & Error Overlay**: Halts writes to `dist/` on diagnostic errors during `pphlx dev` / `pphlx watch` and displays an instant browser error overlay.
- [ ] **Automated CI/CD Verification (`pphlx check --ci`)**: Parallel zero-exit-code AST static analysis and PHP syntax verification for build pipelines.
- [ ] **Official Docker Container Images (`pphlx/compiler`)**: Published Alpine and Debian-slim Docker container images for GitHub Actions, GitLab CI, and Bitbucket Pipelines.
- [ ] **Production Asset Fingerprinting**: Automatic content hashing (e.g. `app.[hash].css`, `app.[hash].js`) and immutable cache header generation.
- [ ] **Compilation Telemetry & Asset Size Audits**: Detailed build metrics, bundle size reporting, and dead-code warnings.
- [ ] **Content Security Policy (CSP) Cryptographic Nonces**: Automatic CSP nonce injection for framework island script tags to pass enterprise security audits.
- [ ] **Serverless & Edge Compilation Targets (`target: "serverless"`)**: Native packaging for AWS Lambda, GCP Cloud Run, and Cloudflare Workers PHP runtimes.
- [ ] **Atomic Zero-Downtime Deployment Packaging**: Production deployment manifest generation for enterprise Nginx/Apache cluster deployments.
- [ ] **Selective Progressive Hydration (`client:visible`, `client:idle`)**: Viewport intersection hydration for production performance optimization.

---

### Phase 3: Enterprise Monolith Generator & Next-Gen Admin Suite — 🚀 Planned
- [ ] **Enterprise App Scaffolder (`pphlx create app <name>`)**: Full-stack enterprise monolith starter scaffolding (Django-like directory structure with native PHP backend models + React/Vue/Svelte islands).
- [ ] **Next-Gen Reactive Auto-Admin Suite (`pphlx admin generate`)**: Instantly generates a modern, glassmorphic reactive Admin UI with real-time filters, global search, CSV export, soft-delete recovery, and role-based permissions (RBAC).
- [ ] **Declarative Security & Permission Directives**: Native RBAC/ABAC role evaluation (`@permission` directive & template guard helpers) with built-in CSRF protection.
- [ ] **Server-Sent Events (SSE) Live Streaming Bridge**: Built-in real-time SSE stream subscriber (`pphlx.stream()`) for dynamic dashboard widgets without WebSocket infrastructure.
- [ ] **Plugin Extension Registry (`pphlx plugin install`)**: Community and enterprise plugin ecosystem for custom AST compiler extensions.

---

## 💬 Community & Feature Requests

Have suggestions for future extensions? Join the discussion on our [GitHub Discussions](https://github.com/pphlx/pphlx/discussions) or open an issue on GitHub!
