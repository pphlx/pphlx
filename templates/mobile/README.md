# PPHLX Native Android Mobile Application

A sovereign, zero-allocation native mobile application for Android, featuring declarative **LX** UI authoring and **Go / Android NDK** backend logic.

## 🚀 Quick Start

1. **Design UI**:
   Open `ui/app.lx` in VS Code with the PPHLX Studio extension or launch PPHLX Studio for visual live preview.

2. **Compile and Build Android APK**:
   ```bash
   pphlx build android
   ```

## 📱 Features
- **Pure Declarative LX**: Fast, concise mobile layout syntax.
- **DO-178C Zero-Allocation Engine**: `malloc = 0` during rendering.
- **Native Android Performance**: Sub-millisecond touch dispatch over `ANativeWindow`.
