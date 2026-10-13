package main

import (
	"fmt"

	pphlx "github.com/pphlx/pphlx"
)

func main() {
	// 1. Initialize PPHLX Native Context with an 8 MB static arena
	ctx, err := pphlx.NewContext(8 * 1024 * 1024)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize PPHLX Native context: %v", err))
	}
	defer ctx.Close()

	// 2. Launch sovereign OS window (1280x800)
	win, err := ctx.CreateWindow("PPHLX Native Desktop Application", 1280, 800)
	if err != nil {
		panic(fmt.Sprintf("Failed to create window: %v", err))
	}

	// 3. Build flexbox UI hierarchy
	root := ctx.Root()
	root.SetBackgroundColor(0xFF0F172A)
	root.SetFlexColumn(pphlx.JustifyCenter, pphlx.AlignCenter, 16)

	// Title
	title := ctx.CreateNode()
	title.SetTextStyle(0xFFFFFFFF, 26, pphlx.TextAlignCenter)
	title.SetBold(true)
	title.SetText("PPHLX Sovereign Native Desktop")
	root.AddChild(title)

	// Subtitle
	sub := ctx.CreateNode()
	sub.SetTextStyle(0xFF94A3B8, 14, pphlx.TextAlignCenter)
	sub.SetText("Zero-Allocation C-ABI • Mathematical Flexbox • Sub-Millisecond Render")
	root.AddChild(sub)

	// Interactive Button
	btn := ctx.CreateButton("Click Me", pphlx.ButtonPrimary)
	btn.SetSizePx(160, 42)
	btn.SetOnClick(func() {
		fmt.Println("[PPHLX] Button clicked on desktop window!")
	})
	root.AddChild(btn)

	// 4. Compute layout and run native OS event pump
	ctx.ComputeLayout(1280, 800)
	win.Run()
}
