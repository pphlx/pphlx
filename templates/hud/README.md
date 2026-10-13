# PPHLX Tactical HUD & Avionics Telemetry

A sovereign, zero-allocation real-time heads-up display (HUD) and instrument telemetry starter powered by the **PPHLX Native C99 Engine** and **Go**.

## 🚀 Quick Start

Run the HUD scope locally:
```bash
go run .
```

Build an optimized release binary:
```bash
go build -ldflags "-H=windowsgui -s -w" -o hud.exe .
```

## 🎨 Visual Layout Editing (.lx)

This starter contains `hud.lx`. You can visually edit, preview, and modify this layout directly inside **PPHLX Studio**:
1. Launch **PPHLX Studio** (`studio.exe`).
2. Click **Import .lx** and choose `hud.lx`.
3. Visually drag and drop telemetry cards, change themes, and live compile to C99 or Go code.

## 🛠️ Features
- **Zero Heap Allocations**: 100% fixed static arena memory. Sub-millisecond rendering for 120 FPS+ game engines and simulators.
- **LX Declarative Syntax**: Clean, human-readable UI definitions compatible with PPHLX Studio.
- **Hardware Blitting**: Direct software rasterization to OS window framebuffers with zero external GUI dependencies.
