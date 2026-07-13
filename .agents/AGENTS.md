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
