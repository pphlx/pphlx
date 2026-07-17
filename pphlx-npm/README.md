# PPHLX Compiler

![npm](https://img.shields.io/npm/v/pphlx)
![npm downloads](https://img.shields.io/npm/dm/pphlx)
![License](https://img.shields.io/npm/l/pphlx)

> Modern web components. Zero runtime. Pure PHP output.

Official Node.js CLI for PPHLX — a high-performance compiler for building modern, component-based PHP applications.

Powered by a Go-compiled WebAssembly (WASI) runtime, PPHLX runs consistently on Windows, macOS, and Linux without requiring native compiler installations.

---

## Features

- **🚀 High-performance compiler**: Powered by Go and compiled natively.
- **⚡ WebAssembly (WASI) powered**: Run anywhere Node.js runs without native compiler dependencies.
- **📦 Zero runtime in production**: The generated output is standalone, with no node modules.
- **🧩 Component-based layout**: Build components using your favorite frontend workflows.
- **🔥 Hot reload**: Fast development server built-in.
- **🌍 Cross-platform**: Instant startup on Windows, macOS, and Linux.
- **🐘 Generates pure PHP**: Easily host your builds on any cheap PHP web hosting server.

---

## Quick Start

Get your project up and running in seconds:

1. **Install PPHLX locally**:
   ```bash
   npm install pphlx
   ```
   *(Note: The installer automatically configures `dev`, `start`, `watch`, `build`, `preview`, and `check` shortcut triggers inside your project's `package.json` scripts block).*

2. **Initialize your project configuration**:
   ```bash
   npx pphlx init
   ```

3. **Start the local development server** (with hot reload and background PHP routing):
   ```bash
   npm run dev
   ```

4. **Verify template diagnostics**:
   ```bash
   npm run check
   ```

5. **Build and preview your production package**:
   ```bash
   npm run build
   npm run preview
   ```

---

## Usage

PPHLX compiles your components into dependency-free, production-ready PHP pages. The generated output requires no Node.js, JavaScript runtime, or PPHLX framework in production.

By default, the compiler looks for a configuration file (`pphlx.config.json` or `pphlx.config.mjs`) in the root of your project directory.

### package.json Scripts

Upon installation, PPHLX automatically configures the following scripts in your project's `package.json` so you do not need to configure them manually:

```json
{
  "scripts": {
    "dev": "pphlx dev",
    "start": "pphlx dev",
    "watch": "pphlx watch",
    "build": "pphlx",
    "preview": "pphlx preview",
    "check": "pphlx check"
  }
}
```

Then simply use the corresponding `npm run <command>`:
- **`npm run dev`**: Launches active hot-reload compilation and dev server.
- **`npm run check`**: Validates template syntax and component imports.
- **`npm run build`**: Compiles layouts and assets into standalone `.php` production files.
- **`npm run preview`**: Starts a local PHP server pointing directly to your compiled production directory.

---

## Project Links

- **Website**: [pphlx.org](https://pphlx.org)
- **GitHub Repository**: [github.com/pphlx/pphlx](https://github.com/pphlx/pphlx)
- **Discord Community**: [Discord Server](https://discord.gg/9ApeZhsG7G)
