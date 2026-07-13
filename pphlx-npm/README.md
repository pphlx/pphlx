# PPHLX Compiler (Node.js/npm Package)

Official Node.js CLI compiler wrapper for **PPHLX**, a high-performance web compiler that builds component-based layouts into standard PHP monolith pages.

This package wraps the Go-compiled WebAssembly (WASI) binary, providing cross-platform execution on Windows, macOS, and Linux out of the box without any native compiler dependencies.

---

## Installation

Install the package locally in your project:
```bash
npm install pphlx
```

Or run it directly using `npx`:
```bash
npx pphlx build
```

---

## Usage

PPHLX uses a configuration file (`pphlx.config.json` or `pphlx.config.mjs`) in the root of your project directory.

### Scripts
Add the build trigger script to your `package.json`:
```json
{
  "scripts": {
    "build": "pphlx build",
    "dev": "pphlx dev"
  }
}
```

Then run:
```bash
npm run build
```
This compiles your layout and components into standard `.php` files inside your output directory.
