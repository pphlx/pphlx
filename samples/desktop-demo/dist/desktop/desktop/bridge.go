package main

import (
	"fmt"
	"runtime"
)

type CustomBridge struct{}

func init() {
	// Register this extension during initialization
	RegisterExtension(func(w DesktopWindow) {
		bridge := &CustomBridge{}
		w.Bind("CustomBridge", bridge)
	})
}

// GetSystemInfo returns host architecture details natively
func (b *CustomBridge) GetSystemInfo() string {
	return fmt.Sprintf("OS: %s, Arch: %s, CPU Cores: %d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}
