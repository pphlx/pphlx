# Changelog

All notable changes to the PPHLX Node.js compiler CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.4-1] - 2026-07-15

### Fixed
- Removed unused experimental WASI module import to eliminate Node.js runtime warning messages.

---

## [1.0.4] - 2026-07-15

### Added
- Integrated automatic local PHP development server (`php -S`) execution directly in the background for `pphlx dev` and `pphlx watch` commands.
- Redesigned development server ready states and startup outputs to mimic the clean Astro CLI logs without emojis.

### Fixed
- Fixed mobile responsive layout padding collapsing inside multiframe dashboards.
- Switched Node.js WASI execution layer to standard `wasm_exec.js` Go WASM runner to resolve Windows filesystem directory-reading (`readdirent: not implemented`) bugs and completely remove experimental WASI startup warnings.

---

## [1.0.3] - 2026-07-14

### Fixed
- Fixed component loading path mappings when compiled on Windows environments.

---

## [1.0.2] - 2026-07-14

### Changed
- Improved hot-reloading file watch engine to throttle rapid successive edits.
- Optimised AST parsing for large template files to reduce compilation latency.

### Fixed
- Fixed compilation of TresJS, SolidJS, and Lit components under specific Windows path scopes.
- Resolved build issue where production output assets contained incomplete asset maps.

---

## [1.0.1] - 2026-04-10

### Added
- Added built-in hot reloading dev server triggers (`pphlx dev`).
- Created project bootstrapping template runner (`pphlx init`).
- Integrated dynamic framework island detection for automatic asset bundle compilation.

### Fixed
- Resolved file locking issues under Windows environments during active rebuild cycles.

---

## [1.0.0] - 2026-02-18

### Added
- Initial public release of the official PPHLX CLI compiler package.
- Embedded WASI compilation core running precompiled Go binaries.
- Side-by-side framework island rendering and server-side PHP data hydration support.
- Production compiler optimization pipeline generating clean, standalone, database-ready PHP assets.
