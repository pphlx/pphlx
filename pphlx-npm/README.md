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

```bash
# Install PPHLX locally
npm install pphlx

# Initialize your project configuration
npx pphlx init

# Start the local development server (with hot reload)
npx pphlx dev

# Build for production deployment
npx pphlx build
```

---

## Usage

PPHLX compiles your components into dependency-free, production-ready PHP pages. The generated output requires no Node.js, JavaScript runtime, or PPHLX framework in production.

By default, the compiler looks for a configuration file (`pphlx.config.json` or `pphlx.config.mjs`) in the root of your project directory.

### package.json Scripts
Add build and development triggers to your project scripts:
```json
{
  "scripts": {
    "dev": "pphlx dev",
    "build": "pphlx build"
  }
}
```

Then run:
```bash
npm run build
```
This compiles your layouts and components into standalone `.php` files inside your output directory.

---

## Project Links

- **Website**: [pphlx.org](https://pphlx.org)
- **GitHub Repository**: [github.com/pphlx/pphlx](https://github.com/pphlx/pphlx)
- **Discord Community**: [Discord Server](https://discord.gg/9ApeZhsG7G)
