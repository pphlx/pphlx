$ErrorActionPreference = "Stop"

$CoreDir = "f:\VS CODE\GO\PPHLX\pphlx-core"
$ReleasesDir = "f:\VS CODE\GO\PPHLX\pphlx-core\releases"

Write-Host "Building PPHLX v1.1.0 Release Tarballs..." -ForegroundColor Cyan

Set-Location $CoreDir

# 1. Darwin arm64
Write-Host "Compiling darwin/arm64..." -ForegroundColor Gray
$env:CGO_ENABLED = "0"
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
$OutPath = Join-Path $ReleasesDir "pphlx"
go build -o $OutPath .
tar -czf (Join-Path $ReleasesDir "pphlx-darwin-arm64.tar.gz") -C $ReleasesDir pphlx

# 2. Darwin amd64
Write-Host "Compiling darwin/amd64..." -ForegroundColor Gray
$env:GOARCH = "amd64"
go build -o $OutPath .
tar -czf (Join-Path $ReleasesDir "pphlx-darwin-amd64.tar.gz") -C $ReleasesDir pphlx

# 3. Linux arm64
Write-Host "Compiling linux/arm64..." -ForegroundColor Gray
$env:GOOS = "linux"
$env:GOARCH = "arm64"
go build -o $OutPath .
tar -czf (Join-Path $ReleasesDir "pphlx-linux-arm64.tar.gz") -C $ReleasesDir pphlx

# 4. Linux amd64
Write-Host "Compiling linux/amd64..." -ForegroundColor Gray
$env:GOARCH = "amd64"
go build -o $OutPath .
tar -czf (Join-Path $ReleasesDir "pphlx-linux-amd64.tar.gz") -C $ReleasesDir pphlx

# 5. Windows amd64
Write-Host "Compiling windows/amd64..." -ForegroundColor Gray
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$WinExe = Join-Path $ReleasesDir "pphlx.exe"
go build -o $WinExe .
Compress-Archive -Path $WinExe -DestinationPath (Join-Path $ReleasesDir "pphlx-windows-amd64.zip") -Force

# Cleanup temp binaries
Remove-Item $OutPath -Force -ErrorAction SilentlyContinue
Remove-Item $WinExe -Force -ErrorAction SilentlyContinue

Write-Host "All v1.1.0 Release Tarballs built successfully!" -ForegroundColor Green
