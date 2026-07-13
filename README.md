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
