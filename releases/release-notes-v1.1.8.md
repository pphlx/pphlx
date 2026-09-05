# PPHLX Compiler v1.1.8 Release Notes

## 🚀 What's New in v1.1.8

### ⚡ Directives Compilation Engine (`set:html`, `set:text`, `is:raw`, `is:inline`)
- **`set:html` Directive**: Supports reactive unescaped HTML content binding (`<div set:html="$rawHtml"></div>` ➔ `<?php echo $rawHtml; ?>`).
- **`set:text` Directive**: Automatically applies XSS escaping via `htmlspecialchars` (`<span set:text="$escapedText"></span>` ➔ `<?php echo htmlspecialchars($escapedText); ?>`).
- **`is:raw` Directive**: Preserves raw code blocks without parsing template expressions or tags (`<code is:raw>...</code>`).
- **`is:inline` Directive**: Preserves `<script is:inline>` and `<style is:inline>` tags directly in place instead of extracting and bundling them into global CSS/JS bundles.

---

## 🔒 SHA256 Checksums

| Archive File | SHA256 Checksum |
| :--- | :--- |
| `pphlx-darwin-arm64.tar.gz` | `74cb87ace8ec1d929cd8a86599b2f1a9c828f7548e44681e4344c4f483ef0cf2` |
| `pphlx-darwin-amd64.tar.gz` | `b33ae76620bce185ae4536d721b39e67878926cce631afd9e90cc75a2fd38dd3` |
| `pphlx-linux-arm64.tar.gz`  | `a7d2a4dc3567a5fecce1c132c6ac18a2f74ff721617d66c34cbed138c2fce23c` |
| `pphlx-linux-amd64.tar.gz`  | `b62abd1d25b77581a6d473f7651d20bac9df097a607110b99bfe723c055945d1` |
| `pphlx-windows-amd64.zip`  | `c8c64d801a548bc4b86159af9b3b897279f0681453bfc548827219b7e79a63bd` |
| `pphlx.msi`                | `1ccca383f3e412d38dd6600e9077fd4dad737c96350d442735819143424a57df` |
