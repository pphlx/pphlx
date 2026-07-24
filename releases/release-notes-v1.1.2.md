# PPHLX v1.1.2 Release Notes

PPHLX `v1.1.2` introduces clean aesthetic starter template scaffolding, embedded binary favicon assets (`favicon.svg` and binary `favicon.ico`), automatic project `README.md` generation with folder structure trees, unified project directory scaffolding, dual-platform verification for Windows & WSL Ubuntu, and `create-pphlx` v1.0.1.

## 🚀 Key Highlights & Enhancements

### 1. Embedded Binary Scaffolder Template
Embedded the complete, modern starter template directly inside the single zero-dependency PPHLX compiler binary (`pphlx-win.exe`, `pphlx-linux`, `pphlx-macos`, `pphlx.wasm`). `pphlx init` and `npm create pphlx@latest` create all required starter files instantly with zero external network downloads.

### 2. Dual Favicon Scaffolding
Running `pphlx init` or `create-pphlx` now automatically generates both `public/favicon.svg` and `public/favicon.ico` (base64 decoded binary ICO) in your project.

### 3. Automated `README.md` Generation
Generated project `README.md` containing the updated folder tree structure, compiler command guide, and documentation links.

### 4. Dual Setup Compatibility (New vs Existing Projects)
- **1-Line New Project Setup**: `npm create pphlx@latest my-app`
- **Manual Existing Project Setup**: `npm install pphlx` -> `npx pphlx init`

---

## 📦 Binary Release Checksums (SHA256)

| Target Platform | Asset Filename | SHA256 Checksum |
| :--- | :--- | :--- |
| **macOS (Apple Silicon)** | `pphlx-darwin-arm64.tar.gz` | `a2f0432e0872775062c4766595cb7d7ef5817fa94ffc14646d253a62ce5706a5` |
| **macOS (Intel x64)** | `pphlx-darwin-amd64.tar.gz` | `fd45fd09796971da37b965e2a4a25a7953533eee3490ae256f64cc59e1122c4d` |
| **Linux (ARM64)** | `pphlx-linux-arm64.tar.gz` | `c2841401838f3bc9c98774ff26594c041596834418bec22ba7f54c59ebb6e134` |
| **Linux (AMD64)** | `pphlx-linux-amd64.tar.gz` | `0e867596167939abf8a2e276867a76349178ff72fa7fc7bce0fc93dabb2faafa` |
| **Windows (AMD64)** | `pphlx-windows-amd64.zip` | `44900dbe91835449c385a450baee81313de1486702bd5581f147bd9f8ea2c2a9` |
