# PPHLX Compiler v1.1.7 Release Notes

## 🚀 What's New in v1.1.7

### 🛠️ Multi-Platform Binary Packaging Sync
Fixed release workflow cross-compilation pipeline to ensure macOS (ARM64/AMD64), Linux (ARM64/AMD64), and Windows binaries packaged inside distribution submodules are synchronized with the compiler version (`v1.1.7`).

### 🔌 MCP ServerInfo Version & Protocol Date
Fixed `pphlx mcp` initialization handler in `main.go` to dynamically return the central `Version` constant (`v1.1.7`) and dynamic release date format (`DD-MM-YYYY`).

---

## 🔒 SHA256 Checksums

| Archive File | SHA256 Checksum |
| :--- | :--- |
| `pphlx-darwin-arm64.tar.gz` | `48635d1b398aa8ca58b355bb619c2338c24415c1f79390c68ff7972bc6267aa1` |
| `pphlx-darwin-amd64.tar.gz` | `43ebf92d371ac1ec7041f5a62aabc13f5261223facbe3b1af52e66bf724b03e0` |
| `pphlx-linux-arm64.tar.gz`  | `ac5db1e70ff7c9203394037ae4c53d4207e618bae7d15d8ecce638b833a7f72f` |
| `pphlx-linux-amd64.tar.gz`  | `4485f06753e1a47c740aaba278a87605fd388111c4681ea8a0caa985d33b6e2c` |
| `pphlx-windows-amd64.zip`  | `33cfa4b30ddebdad19b6ebeb391b33c30beefcd8be720ac0f578dccfb4648f39` |
| `pphlx.msi`                | `881ac36a0058bc53d756d7a6c4c1a16137f6248f6754e6b06cb9b2725825f11b` |
