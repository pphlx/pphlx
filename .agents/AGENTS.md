# PPHLX Compiler & Component Syntax Guidelines

## 1. HTML Attribute Tokenization with PHP
*   Do not parse HTML attributes using simple negative characters sets like `[^>]*`. PHP templates commonly embed variables inside attributes (e.g., `href="<?php echo $url; ?>"`). The closing tag `?>` of the PHP block will prematurely close the attribute parser.
*   **Rule:** Always tokenise attributes using patterns that explicitly skip over quoted string literals (`"..."`, `'...'`) and PHP blocks (`<\?php.*?\?>`), such as:
    `attrPattern := (\s+(?:[^>"]|"[^"]*"|'[^']*'|<\?php.*?\?>)*)`

## 2. Compiler System Placeholders
*   **Rule:** When implementing template interpolation for variables (like `{{title}}`), always verify that system asset anchors (like `{{PPHLX_CSS}}` and `{{PPHLX_JS}}`) are skipped and preserved for the final asset injection stage.

## 3. Configuration Architecture
*   **Rule:** Maintain a strict separation of configuration files:
    *   `pphlx.json` (Project Manifest): Holds project metadata, script triggers (e.g., `"build": "pphlx build"`), and dependency declarations.
    *   `pphlx.config.json` (Compiler config): Holds specific build configurations (e.g., `srcDir`, `outDir`, `cssOut`, and `jsOut`).

## 4. Repository & Release Target Architecture
*   **Workspace Repository Isolation**:
    *   The root workspace repository (`f:\VS CODE\GO\PPHLX`) is a private development environment. Do not push intermediate documentation commits, changelog updates, or release-specific tag files directly to the private root. Keep its commit log matching core feature updates.
    *   All package manifests, changelogs, and binary distribution builds belong in their respective public subdirectories (`pphlx-npm`, `pphlx-composer`, `homebrew-tap`).
    *   Release tags and tarballs intended for public distribution must be pushed directly to the public release repository at `https://github.com/pphlx/pphlx.git`, rather than the private origin remote (`KillerTyzon/pphlx.git`).

## 5. System Environments & Antivirus Alerts
*   **No Engine Compromises**: Do not modify PPHLX engine logic or write runtime workarounds to bypass local operating system or antivirus security alerts. Design the CLI commands and download pipelines exactly how they are supposed to function natively; the user will configure antivirus exclusions or settings as needed.

