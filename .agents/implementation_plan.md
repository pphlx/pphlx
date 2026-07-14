# Playbooks Asset Resolution & Client-Side IDE Compilation Plan

This plan establishes the architecture to resolve missing CSS/JS asset paths inside the Play IDE preview iframe, and defines how client-side edits are compiled and exported on-the-fly.

## Proposed Architecture

To address the asset path conflicts (`/assets/css/app.css` and `/assets/js/app.js` returning 404s inside the sandbox iframe), we will implement a dual-phase path rewriting and client-side VFS compilation system.

### Phase 1: Playbook Asset Path Rewriting
When loading precompiled playbooks from `/playbooks/{playbook-name}/`, any references to relative assets (e.g. `assets/css/app.css` or `assets/js/pphlx_vite.js`) in the generated HTML must map back to their respective playbook folder in the public workspace directory: `/playbooks/{playbook-name}/dist/assets/`.

We will modify the preview iframe generator inside `play.astro` to perform dynamic string rewriting:
* Matches any `href="assets/"` or `src="assets/"` patterns.
* Prefixes them with the active playbook's path: `/playbooks/${activePlaybook}/dist/assets/`.

---

### Phase 2: Client-Side Compilation & Virtual File System (VFS)
To allow users to modify files on the fly without making server roundtrips:
1. **Monaco VFS Synchronization**:
   * All file modifications inside the Monaco Editor are immediately saved to a local state variable `files` (mapping file path -> content string) in the browser.
2. **On-the-Fly HTML/CSS Parser**:
   * The client-side parser reads layout files and expands `@import` component placeholders directly using the virtual memory copy, refreshing the preview iframe content on every keystroke.
3. **Standalone Zip Exporter**:
   * When the user clicks the **Download** button in the Play IDE toolbar, we will load `JSZip` from a CDN and pack all files in the browser VFS (including the compiled `dist/index.php` and assets) into a zip file, triggering a client-side download.

---

## Proposed Changes

### Play IDE Page

#### [MODIFY] [play.astro](file:///f:/VS%20CODE/GO/PPHLX/pphlx-org/src/pages/play.astro)
* Update `doCompile()` to dynamically rewrite asset links relative to the active playbook directory.
* Integrate JSZip to package files from the virtual files system into a downloadable standalone zip file when the download trigger is clicked.

---

## Verification Plan

### Manual Verification
1. **Playbook Preview Verification**:
   * Open the Play IDE for the `multiframe-dashboard` playbook.
   * Verify that the preview frame loads all styles and scripts successfully (returning `200 OK` instead of `404 Not Found`).
2. **On-The-Fly Compilation Verification**:
   * Edit text in the HTML/PHP layouts and verify that the preview frame re-renders the changes instantly.
3. **Zip Download Verification**:
   * Click the download button in the toolbar and verify it triggers a download of a complete ZIP archive containing the standalone layout and assets.
