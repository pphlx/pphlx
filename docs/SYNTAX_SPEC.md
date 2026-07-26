# The `.pphx` File Format - Language Syntax Specification

**Version:** 1.0  
**Status:** Official Standard  
**Date:** 2026-07-26  

---

## Table of Contents

1. [File Structure](#1-file-structure)
2. [PPHLX Core Syntax Tokens](#2-pphlx-core-syntax-tokens)
   - 2.1 PHP Echo Expression Tag (`{|= $expr |}`)
   - 2.2 PHP Code Statement Block (`{| $stmt |}`)
   - 2.3 Component & Layout Imports (`@import`)
3. [Multi-Framework Island Components](#3-multi-framework-island-components)
   - 3.1 Supported Frameworks
   - 3.2 Client Hydration Directives (`client:load`, `client:visible`, `client:idle`)
   - 3.3 Component Property Serialization Invariant (`json_encode`)
4. [Autonomous Asset Management](#4-autonomous-asset-management)
   - 4.1 CSS Bundle Injection & 0-Byte File Omission
   - 4.2 JS Virtual Endpoint Streaming
5. [Layouts & Template Composition](#5-layouts--template-composition)

---

## 1. File Structure

A `.pphx` file is a PPHLX single-file template containing component imports, PHP database mock state, HTML structure, and island component declarations.

```pphx
@import Layout from 'layouts/Layout.pphx'
@import ProductImage from 'components/ProductImage.js'
@import StarRating from 'components/StarRating.solid.jsx'
@import BuyButton from 'components/BuyButton.svelte'

{|
$productName = "MacBook Pro";
$priceUSD = 2599.00;
$initialStars = 5;
|}

<Layout>
  <h2>{|= $productName |}</h2>
  <StarRating initialRating={|= $initialStars |} client:load />
  <BuyButton price={|= $priceUSD |} client:load />
</Layout>
```

---

## 2. PPHLX Core Syntax Tokens

### 2.1 PHP Echo Expression Tag (`{|= $expr |}`)

Outputs evaluated PHP expressions directly into HTML text nodes or component property values.

- **PPHLX Token**: `{|= $expression |}`
- **Compiled PHP Output**: `<?php echo $expression; ?>`
- **Example**:
  ```pphx
  <p>Price: ${|= number_format($priceUSD, 2) |}</p>
  ```

### 2.2 PHP Code Statement Block (`{| $stmt |}`)

Executes raw PHP server-side code blocks (variable assignments, database queries, session checks).

- **PPHLX Token**: `{| $statement |}`
- **Compiled PHP Output**: `<?php $statement ?>`
- **Example**:
  ```pphx
  {|
  $user = getCurrentUser();
  $stockLeft = 12;
  |}
  ```

### 2.3 Component & Layout Imports (`@import`)

Declares imported layout templates, React/Vue/Svelte/SolidJS/Preact components, or child PPHLX components at the top of the file.

- **Syntax**: `@import ComponentName from 'path/to/component'`
- **Supported File Extensions**:
  - `.pphx` — PPHLX template component
  - `.js` / `.jsx` — React or Preact component
  - `.vue` — Vue 3 Single-File Component (SFC)
  - `.svelte` — Svelte 4 / Svelte 5 component
  - `.solid.jsx` / `.solid.tsx` — SolidJS component
  - `.ts` / `.tsx` — TypeScript component

---

## 3. Multi-Framework Island Components

PPHLX supports concurrent execution and client-side hydration of multi-framework UI islands (React, Vue, Svelte, SolidJS, Preact) on a single PHP monolithic thread.

### 3.1 Client Hydration Directives

- `client:load` — Hydrate component immediately on page load.
- `client:visible` — Hydrate component when scrolled into viewport (via IntersectionObserver).
- `client:idle` — Hydrate component when browser main thread is idle (via requestIdleCallback).

### 3.2 Component Property Serialization Invariant

When passing PHP variables or `{|= $expr |}` tags into island component attributes:

```pphx
<FeedbackCard title="{|= $reactTitle |}" client:load />
```

The PPHLX compiler extracts the PHP expression `$reactTitle` and serializes it as `json_encode($reactTitle)` inside the island script container:

```html
<div id="pphlx-feedbackcard-3acf88c3" class="pphlx-island" data-component="FeedbackCard" data-framework="react" data-hydrate="load"></div>
<script>
  window.pphlxProps = window.pphlxProps || {};
  window.pphlxProps["pphlx-feedbackcard-3acf88c3"] = {"title": <?php echo json_encode($reactTitle); ?>};
</script>
```

When executed on the server, PHP evaluates the expression directly into valid JSON, preventing string-escaping errors and enabling 100% type-safe hydration in client frameworks.

---

## 4. Autonomous Asset Management

### 4.1 CSS Bundle Injection & 0-Byte File Omission

- If CSS rules exist in the project, `<link rel="stylesheet" href="assets/css/app.css">` is autonomously injected before `</head>`.
- If **0 bytes of custom CSS** are generated, PPHLX **omits creating empty `app.css` files or empty `css/` directories** and does NOT inject `<link rel="stylesheet">`.

### 4.2 JS Virtual Endpoint Streaming

- `<script src="assets/js/app.js"></script>` is injected before `</body>`.
- In `pphlx dev`, `/assets/js/app.js` streams virtual JS code directly from RAM with `Cache-Control: no-cache, no-store, must-revalidate` headers.

---

## 5. Layouts & Template Composition

PPHLX templates wrap content using layout tags (`<Layout>...</Layout>`). The inner body content replaces the `{{slot}}` or `{|= $slot |}` placeholder in the layout template.
