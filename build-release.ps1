$ErrorActionPreference = "Stop"

$CoreDir = Get-Location
$RootDir = Split-Path $CoreDir -Parent
$ReleasesDir = Join-Path $CoreDir "releases"

if (!(Test-Path $ReleasesDir)) {
    New-Item -ItemType Directory -Force -Path $ReleasesDir | Out-Null
}

# 1. Detect Release Version dynamically from main.go (Single Source of Truth)
$MainGoPath = Join-Path $CoreDir "main.go"
if (!(Test-Path $MainGoPath)) {
    Write-Error "Could not find main.go at $MainGoPath to detect release version."
    exit 1
}

$VersionLine = Get-Content $MainGoPath | Where-Object { $_ -match 'const\s+Version\s*=\s*"([^"]+)"' } | Select-Object -First 1
if ($VersionLine -match 'const\s+Version\s*=\s*"([^"]+)"') {
    $Version = $Matches[1]
} else {
    Write-Error "Failed to parse 'const Version' from $MainGoPath!"
    exit 1
}

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " Building PPHLX v$Version Multi-Platform Release" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# 2. Cross-Compile All Go Binaries & WebAssembly
$env:CGO_ENABLED = "0"
Write-Host "`n[Step 1/5] Cross-compiling Go binaries for all platform targets..." -ForegroundColor Yellow

# Windows x64
Write-Host "  -> Compiling windows/amd64..." -ForegroundColor Gray
$env:GOOS = "windows"; $env:GOARCH = "amd64"
$WinExe = Join-Path $CoreDir "pphlx-win.exe"
go build -o $WinExe main.go main_cli.go

# Linux x64 & ARM64
Write-Host "  -> Compiling linux/amd64..." -ForegroundColor Gray
$env:GOOS = "linux"; $env:GOARCH = "amd64"
$LinuxAmd64 = Join-Path $CoreDir "pphlx-linux-amd64"
go build -o $LinuxAmd64 main.go main_cli.go
$LinuxBin = Join-Path $CoreDir "pphlx-linux"
Copy-Item $LinuxAmd64 $LinuxBin -Force

Write-Host "  -> Compiling linux/arm64..." -ForegroundColor Gray
$env:GOOS = "linux"; $env:GOARCH = "arm64"
$LinuxArm64 = Join-Path $CoreDir "pphlx-linux-arm64"
go build -o $LinuxArm64 main.go main_cli.go

# macOS Intel & Apple Silicon
Write-Host "  -> Compiling darwin/amd64..." -ForegroundColor Gray
$env:GOOS = "darwin"; $env:GOARCH = "amd64"
$MacosAmd64 = Join-Path $CoreDir "pphlx-macos-amd64"
go build -o $MacosAmd64 main.go main_cli.go

Write-Host "  -> Compiling darwin/arm64..." -ForegroundColor Gray
$env:GOOS = "darwin"; $env:GOARCH = "arm64"
$MacosArm64 = Join-Path $CoreDir "pphlx-macos-arm64"
go build -o $MacosArm64 main.go main_cli.go
$MacosBin = Join-Path $CoreDir "pphlx-macos"
Copy-Item $MacosArm64 $MacosBin -Force

# WebAssembly
Write-Host "  -> Compiling js/wasm..." -ForegroundColor Gray
$env:GOOS = "js"; $env:GOARCH = "wasm"
$WasmBin = Join-Path $CoreDir "pphlx.wasm"
go build -o $WasmBin main.go main_cli.go

# Reset Environment Variables
Remove-Item env:GOOS -ErrorAction SilentlyContinue
Remove-Item env:GOARCH -ErrorAction SilentlyContinue

# 3. Create Tarballs & Zip Archives inside releases/
Write-Host "`n[Step 2/5] Creating release archives inside releases/..." -ForegroundColor Yellow
tar -czf (Join-Path $ReleasesDir "pphlx-darwin-arm64.tar.gz") -C $CoreDir pphlx-macos
tar -czf (Join-Path $ReleasesDir "pphlx-darwin-amd64.tar.gz") -C $CoreDir pphlx-macos-amd64
tar -czf (Join-Path $ReleasesDir "pphlx-linux-arm64.tar.gz") -C $CoreDir pphlx-linux-arm64
tar -czf (Join-Path $ReleasesDir "pphlx-linux-amd64.tar.gz") -C $CoreDir pphlx-linux-amd64
Compress-Archive -Path $WinExe -DestinationPath (Join-Path $ReleasesDir "pphlx-windows-amd64.zip") -Force

# 4. Build MSI Installer
Write-Host "`n[Step 3/5] Building Windows MSI installer..." -ForegroundColor Yellow
if (Test-Path (Join-Path $CoreDir "build-msi.ps1")) {
    & (Join-Path $CoreDir "build-msi.ps1")
}

# 5. AUTOMATED DISTRIBUTION SYNCING
Write-Host "`n[Step 4/5] Syncing fresh binaries to submodules (npm, composer, website)..." -ForegroundColor Yellow

$NpmBin = Join-Path $RootDir "pphlx-npm\bin"
$ComposerBin = Join-Path $RootDir "pphlx-composer\bin"
$OrgPublic = Join-Path $RootDir "pphlx-org\public"

if (Test-Path $NpmBin) {
    Copy-Item $WinExe "$NpmBin\pphlx-win.exe" -Force
    Copy-Item $LinuxBin "$NpmBin\pphlx-linux" -Force
    Copy-Item $LinuxAmd64 "$NpmBin\pphlx-linux-amd64" -Force
    Copy-Item $LinuxArm64 "$NpmBin\pphlx-linux-arm64" -Force
    Copy-Item $MacosBin "$NpmBin\pphlx-macos" -Force
    Copy-Item $MacosAmd64 "$NpmBin\pphlx-macos-amd64" -Force
    Copy-Item $MacosArm64 "$NpmBin\pphlx-macos-arm64" -Force
    Copy-Item $WasmBin "$NpmBin\pphlx.wasm" -Force
    Write-Host "  [OK] Synced 8 binary targets to pphlx-npm/bin/" -ForegroundColor Green
}

if (Test-Path $ComposerBin) {
    Copy-Item $WinExe "$ComposerBin\pphlx-win.exe" -Force
    Copy-Item $LinuxBin "$ComposerBin\pphlx-linux" -Force
    Copy-Item $MacosBin "$ComposerBin\pphlx-macos" -Force
    Copy-Item $WasmBin "$ComposerBin\pphlx.wasm" -Force
    Write-Host "  [OK] Synced 4 binary targets to pphlx-composer/bin/" -ForegroundColor Green
}

if (Test-Path $OrgPublic) {
    Copy-Item $WasmBin "$OrgPublic\pphlx.wasm" -Force
    if (Test-Path "$CoreDir\install.ps1") { Copy-Item "$CoreDir\install.ps1" "$OrgPublic\install.ps1" -Force }
    if (Test-Path "$CoreDir\install.sh") { Copy-Item "$CoreDir\install.sh" "$OrgPublic\install.sh" -Force }
    Write-Host "  [OK] Synced public assets to pphlx-org/public/" -ForegroundColor Green
}

# 6. Display Checksums
Write-Host "`n[Step 5/5] Computed SHA256 Checksums for Release Notes:" -ForegroundColor Yellow
Get-ChildItem "$ReleasesDir\*" -Include *.tar.gz, *.zip, *.msi | ForEach-Object {
    $hash = (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower()
    Write-Host "  $($_.Name): $hash" -ForegroundColor Gray
}

Write-Host "`n=================================================" -ForegroundColor Green
Write-Host " Build and Automated Distribution Sync Complete (v$Version)!" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
