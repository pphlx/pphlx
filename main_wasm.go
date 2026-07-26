//go:build js && wasm

package main

import (
	"path/filepath"
	"syscall/js"
)

func main() {
	js.Global().Set("compilePphlxNative", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return map[string]any{"php": "", "error": "missing arguments"}
		}
		
		content := args[0].String()
		
		// Parse VFS if provided
		if len(args) > 1 && args[1].Type() == js.TypeObject {
			jsVfs := args[1]
			vfs := make(map[string]string)
			keys := js.Global().Get("Object").Call("keys", jsVfs)
			for i := 0; i < keys.Length(); i++ {
				key := keys.Index(i).String()
				vfs[key] = jsVfs.Get(key).String()
			}
			VirtualFiles = vfs
		} else {
			VirtualFiles = nil
		}
		
		currentFile := ""
		if len(args) > 2 {
			currentFile = args[2].String()
		}
		
		currentDir := filepath.Dir(currentFile)
		if currentDir == "." {
			currentDir = ""
		}
		
		// Single source of truth: Call CompilePageWithAssets defined in main.go
		compiledPHP, css, jsList, err := CompilePageWithAssets(content, currentDir, "src")
		
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		
		// Flatten compiled CSS and JS bundles
		cssBundle := ""
		for _, c := range css {
			cssBundle += c + "\n"
		}
		jsBundle := ""
		for _, j := range jsList {
			jsBundle += j + "\n"
		}
		
		return map[string]any{
			"php":   compiledPHP,
			"css":   cssBundle,
			"js":    jsBundle,
			"error": errStr,
		}
	}))
	
	// Keep Go WebAssembly VM alive
	select {}
}
