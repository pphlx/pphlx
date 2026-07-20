# PPHLX Compiler Core

PPHLX (`pphlx.org`) is a fast, zero-dependency compiler written in Go that compiles modern, component-based PHP templates (`.pphx`) into standard, standalone, production-ready `.php` files.

It gives developers the speed, layout nesting, and component reusability of modern frontend frameworks (like React or Astro) while deploying native, zero-runtime overhead PHP pages compatible with any environment (including WHMCS, WordPress, or standard shared hosting).

---

## Key Features

*   **Brace-Pipe Delimiter Syntax:** Ergonomic, collision-free `{|= ... |}` (echo value) and `{| ... |}` (logic block) syntax for writing server-side PHP expressions directly in HTML and framework attributes.
*   **Multi-Framework Islands Hydration:** Seamlessly mount and run **React, Preact, SolidJS, Vue 3, and Svelte 4** interactive components side-by-side on the same PHP page.
*   **Vite Compiler Delegation:** Automatically detects Svelte, Vue, and SolidJS files and delegates bundling to a local, sub-second Vite compile pass.
*   **Recursive Component Nesting:** Build templates (like `Card.pphx` or `Button.pphx`) and nest them infinitely within each other.
*   **Arbitrary Layout Wrapping:** Create independent layouts (e.g., `AdminLayout.pphx`) and wrap pages dynamically (`<AdminLayout>...</AdminLayout>`).
*   **Asset Extraction & Bundling:** Automatically extracts inline `<style>` and `<script>` blocks from components and aggregates them into unified `app.css` and `app.js` bundles.
*   **Recursive File Watcher (`dev`/`watch` mode):** Instantly recompiles output directories on save, featuring built-in event debouncing.
*   **Zero Production Runtime:** Output files are 100% standard PHP. No database hooks, Node server configurations, or background processes required in production.

---

## Template Syntax Reference

PPHLX provides a brand-aligned, zero-typing-fatigue syntax that bridges frontend markup and backend PHP logic:

### 1. Echo Expression: `{|= [expression] |}`
Compiles directly to standard PHP echo syntax. Safe to use in text nodes and attributes:
```html
<h1>Welcome, {|= $user->name |}</h1>
<PreactCart price="{|= $productPrice |}" client:load />
```
*Output:*
```html
<h1>Welcome, <?php echo $user->name; ?></h1>
<div id="pphlx-preactcart-..." class="pphlx-island"></div>
<script>window.__PPHLX_DATA__["..."] = {"price": "<?php echo $productPrice; ?>"};</script>
```

### 2. Logic & Statements: `{| [statement] |}`
Compiles directly to standard PHP control/logical syntax blocks:
```html
{| if ($isAdmin): |}
  <AdminDashboard client:load />
{| endif; |}
```

---

## Configuration Architecture

PPHLX uses a dual-configuration structure:

### 1. Project Manifest (`pphlx.json`)
Analogous to `package.json`. Declares project meta, triggers, and third-party UI packages:
```json
{
  "name": "my-portal-addon",
  "version": "1.0.0",
  "scripts": {
    "dev": "pphlx dev",
    "build": "pphlx build"
  },
  "dependencies": {
    "pphlx-ui-core": "^1.0.0"
  }
}
```

### 2. Compiler Config (`pphlx.config.mjs`)
Analogous to `astro.config.mjs`. Configures source directories, target paths, and asset outputs:
```javascript
import { defineConfig } from "pphlx/config";

export default defineConfig({
  srcDir: "src",
  outDir: "dist",
  cssOut: "dist/assets/css/app.css",
  jsOut: "dist/assets/js/app.js"
});
```

---

## Multi-Target Compilation

PPHLX is a multi-target application engine. By default, it compiles `.pphx` files to standard, production-ready `.php` files. You can configure alternative compilation targets inside `pphlx.config.json` or override them on the fly using CLI flags.

### Supported Targets
1.  **`"php"`** *(Default)*: Compiles templates into dynamic `.php` files for standard web servers.
2.  **`"standalone"`**: Packages all compiled files and static assets into a single, headless executable Go binary with an embedded routing server (`app` or `app.exe`).
3.  **`"desktop"`**: Compiles the codebase into an installable native desktop application using a GPU-accelerated WebView (utilizing pure-Go `webview2` on Windows for zero-CGO builds, and `webview_go` on macOS/Linux).
4.  **`"android"`**: Natively scaffolds a complete Gradle Android Studio project structure inside `dist/android/` with WebView clients and preloaded static assets.
5.  **`"ios"`**: Natively scaffolds a Swift-based Xcode project structure inside `dist/ios/` with WKWebView controllers and preloaded static assets.
6.  **`"ssg"`**: Compiles the codebase to static, raw `.html` (evaluating PHP blocks to static content at build time).
7.  **`"blade"` / `"twig"`**: Translates templates to Laravel-native Blade or Symfony-native Twig views.

---

## Desktop Native App Engine

When targeting `"desktop"`, PPHLX injects the `pphlx.desktop` API directly into your Javascript runtime, allowing your web layout files (React, Svelte, Vue, or raw HTML) to interface with the operating system:

### 1. Core Native Drivers
*   `pphlx.desktop.openFileDialog()`: Opens the native OS file picker and returns the selected filepath.
*   `pphlx.desktop.saveFileDialog()`: Opens the native OS save file dialog.
*   `pphlx.desktop.showNotification(title, message)`: Triggers a native system alert/toast notification.
*   `pphlx.desktop.window.close()`: Gracefully terminates the desktop app process.

### 2. Custom Go Bridge Extensions
For custom hardware integration (e.g. barcode scanners, card readers) or low-level performance code, developers can write local Go files inside `src/desktop/` (e.g. `src/desktop/bridge.go`):

```go
package main

import "fmt"

type CustomBridge struct{}

func init() {
    // Register this bridge extension during app boot
    RegisterExtension(func(w DesktopWindow) {
        bridge := &CustomBridge{}
        w.Bind("CustomBridge", bridge) // Binds window.CustomBridge in JS
    })
}

func (b *CustomBridge) CustomHardwareAction(port string) string {
    return fmt.Sprintf("Interfaced with port: %s", port)
}
```

---

## CLI Usage

PPHLX runs locally on your development machine as a single native binary:

### Compile Once
```bash
pphlx
# or
pphlx build
```

### Start Watcher (Rebuild on Save)
```bash
pphlx dev
# or
pphlx watch
```

---

## License
Licensed under the [MIT License](LICENSE).
