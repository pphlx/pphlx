# PPHLX Native Desktop Application

A lightweight, zero-allocation native desktop application for Windows, macOS, and Linux powered by the **PPHLX Native C99 Engine** and **Go**.

## 🚀 Quick Start

Run the application locally:
```bash
go run .
```

Build a release binary:
```bash
go build -ldflags "-H=windowsgui -s -w" -o app.exe .
```

## 🛠️ Features
- **Zero-Allocation Layout**: Mathematical flexbox layout with 0 heap mallocs during rendering.
- **Sovereign Windowing**: High-performance OS windowing with native hardware blitting.
- **Pure Go Bindings**: Clean, expressive Go API over the PPHLX C-ABI.
