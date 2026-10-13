package main

import (
	"fmt"

	pphlx "github.com/pphlx/pphlx"
)

func main() {
	// 1. Initialize PPHLX Context with a 4 MB static telemetry arena
	ctx, err := pphlx.NewContext(4 * 1024 * 1024)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize PPHLX Native context: %v", err))
	}
	defer ctx.Close()

	// 2. Open HUD Scope window (800x480 Tactical Instrument Display)
	win, err := ctx.CreateWindow("PPHLX Tactical HUD & Avionics Telemetry", 800, 480)
	if err != nil {
		panic(fmt.Sprintf("Failed to create HUD window: %v", err))
	}

	// 3. Construct Root HUD Canvas
	root := ctx.Root()
	root.SetBackgroundColor(0xFF050B14)
	root.SetFlexColumn(pphlx.JustifyBetween, pphlx.AlignStretch, 16)
	root.SetPadding(16, 16, 16, 16)

	// Top Status Ribbon
	statusBar := ctx.CreateNode()
	statusBar.SetFlexRow(pphlx.JustifyBetween, pphlx.AlignCenter, 12)

	sysBadge := ctx.CreateBadge("SYSTEM: ARMED & READY", 0xFF131D2E, 0xFF34D399)
	statusBar.AddChild(sysBadge)

	statusGroup := ctx.CreateNode()
	statusGroup.SetFlexRow(pphlx.JustifyEnd, pphlx.AlignCenter, 16)

	radarTxt := ctx.CreateText("RADAR: ACTIVE", 11, 0xFF38BDF8)
	radarTxt.SetBold(true)
	statusGroup.AddChild(radarTxt)

	gpsTxt := ctx.CreateText("GPS: LOCKED", 11, 0xFF34D399)
	gpsTxt.SetBold(true)
	statusGroup.AddChild(gpsTxt)

	statusBar.AddChild(statusGroup)
	root.AddChild(statusBar)

	// Telemetry Pods Row (Airspeed, PFD, Altitude)
	podsRow := ctx.CreateNode()
	podsRow.SetFlexRow(pphlx.JustifyBetween, pphlx.AlignCenter, 16)

	// Pod 1: Airspeed
	iasCard := ctx.CreateCard(230, 360, 0xFF0F172A, 0xFF1E293D, 12)
	iasCard.SetFlexColumn(pphlx.JustifyStart, pphlx.AlignStart, 8)
	iasCard.SetPadding(16, 16, 16, 16)

	iasTitle := ctx.CreateText("AIRSPEED (IAS)", 13, 0xFF38BDF8)
	iasTitle.SetBold(true)
	iasCard.AddChild(iasTitle)

	iasVal := ctx.CreateText("450.0 KTS", 28, 0xFFFFFFFF)
	iasVal.SetBold(true)
	iasCard.AddChild(iasVal)

	iasSub := ctx.CreateText("Mach 0.68 • Ground 482 KTS", 11, 0xFF64748B)
	iasCard.AddChild(iasSub)
	podsRow.AddChild(iasCard)

	// Pod 2: Primary Flight Display (PFD)
	pfdCard := ctx.CreateCard(260, 360, 0xFF0A0F1A, 0xFF0284C7, 16)
	pfdCard.SetFlexColumn(pphlx.JustifyCenter, pphlx.AlignCenter, 12)
	pfdCard.SetPadding(16, 16, 16, 16)

	pfdBadge := ctx.CreateBadge("PRIMARY FLIGHT DISPLAY", 0xFF0C4A6E, 0xFF38BDF8)
	pfdCard.AddChild(pfdBadge)

	pfdTitle := ctx.CreateText("Level Flight", 20, 0xFFFFFFFF)
	pfdTitle.SetBold(true)
	pfdCard.AddChild(pfdTitle)

	pfdSub := ctx.CreateText("Pitch 0.0° • Roll 0.0°", 12, 0xFF94A3B8)
	pfdCard.AddChild(pfdSub)
	podsRow.AddChild(pfdCard)

	// Pod 3: Altitude
	altCard := ctx.CreateCard(230, 360, 0xFF0F172A, 0xFF1E293D, 12)
	altCard.SetFlexColumn(pphlx.JustifyStart, pphlx.AlignStart, 8)
	altCard.SetPadding(16, 16, 16, 16)

	altTitle := ctx.CreateText("ALTITUDE (ALT)", 13, 0xFF38BDF8)
	altTitle.SetBold(true)
	altCard.AddChild(altTitle)

	altVal := ctx.CreateText("32,000 FT", 28, 0xFFFFFFFF)
	altVal.SetBold(true)
	altCard.AddChild(altVal)

	altSub := ctx.CreateText("Baro 29.92 InHg • V/S 0 FPM", 11, 0xFF64748B)
	altCard.AddChild(altSub)
	podsRow.AddChild(altCard)

	root.AddChild(podsRow)

	// 4. Compute layout and pump native hardware blit loop
	ctx.ComputeLayout(800, 480)
	fmt.Println("[PPHLX HUD] Flight telemetry initialized with 0 dynamic heap allocations.")
	win.Run()
}
