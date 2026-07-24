# PPHLX v1.1.3 Release Notes

PPHLX `v1.1.3` introduces smart dev server HTTP routing, automatic MIME-type detection (`text/html; charset=utf-8` for `.php` files), static `public/` directory asset copying (`favicon.svg` and binary `favicon.ico`), official cyan/white flame SVG branding, and cross-platform dev server compatibility across Windows and WSL Ubuntu.

## 🚀 Key Highlights & Enhancements

### 1. Smart Dev Server HTTP Routing & MIME-Type Polyfill
Updated `pphlx dev` server handler to automatically route `/` and `/index.php` to render `text/html; charset=utf-8` directly in all browsers (preventing directory listings and accidental file downloads on systems without a PHP CLI installed).

### 2. Static `public/` Directory Copying Engine
`pphlx build` and `pphlx dev` now automatically walk the project `public/` directory and copy all static assets (`public/favicon.svg`, `public/favicon.ico`, `public/robots.txt`) directly into `outDir` (`dist/`), ensuring zero 404 errors for web root assets.

### 3. Official PPHLX Flame Favicon Branding
Scaffolding engines (`pphlx init` and `create-pphlx`) now generate `public/favicon.svg` matching the official cyan/white PPHLX flame SVG design with dark-mode media query support (`prefers-color-scheme: dark`).

---

## 📦 Binary Release Checksums (SHA256)

| Target Platform | Asset Filename | SHA256 Checksum |
| :--- | :--- | :--- |
| **macOS (Apple Silicon)** | `pphlx-darwin-arm64.tar.gz` | `6e1f2ab04bee5bb6a5ee26ad67f3b2e145e0f54df0677bd9fac1ba667b609031` |
| **macOS (Intel x64)** | `pphlx-darwin-amd64.tar.gz` | `4409143488e2998f104d4b73ca39e850de9001298052e250da8e4e0affd1c2df` |
| **Linux (ARM64)** | `pphlx-linux-arm64.tar.gz` | `1aa2c67ee839cf73cdd1fd5f033dc9fbc42812f7cd760439ea8e8a76448e48c5` |
| **Linux (AMD64)** | `pphlx-linux-amd64.tar.gz` | `844b1d7f0d2f97259d7b9d5d461689a5c3927fe0de7b27a89264cc18c8a5a9bf` |
| **Windows (AMD64)** | `pphlx-windows-amd64.zip` | `53f91081324c8136890cb330f1c1ab3658a7bf12563c5c35f78286cd0a34e244` |
