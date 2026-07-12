# Piplex MSI Installer Auto-Builder Script
# This script downloads WiX Toolset binaries temporarily and compiles the piplex.msi installer.

$ProjectDir = Get-Location
$TempWixDir = Join-Path $ProjectDir ".wix-bin"

# 1. Compile the Go binary first to ensure it's up to date
Write-Host "Compiling Go binary..." -ForegroundColor Cyan
go build -o piplex.exe main.go
if ($LASTEXITCODE -ne 0) {
    Write-Error "Go compilation failed!"
    exit 1
}

# 2. Check if WiX compiler (candle/light) is installed on system
$CandlePath = Get-Command candle -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source
$LightPath = Get-Command light -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source

if (-not $CandlePath -or -not $LightPath) {
    Write-Host "WiX Toolset not found on PATH. Downloading portable WiX binaries..." -ForegroundColor Yellow
    if (-not (Test-Path $TempWixDir)) {
        New-Item -ItemType Directory -Force -Path $TempWixDir | Out-Null
    }
    
    $WixZipUrl = "https://github.com/wixtoolset/wix3/releases/download/wix3112rtm/wix311-binaries.zip"
    $ZipPath = Join-Path $TempWixDir "wix-binaries.zip"
    
    if (-not (Test-Path $ZipPath)) {
        Write-Host "Downloading: $WixZipUrl" -ForegroundColor Gray
        Invoke-WebRequest -Uri $WixZipUrl -OutFile $ZipPath -UseBasicParsing
    }
    
    Write-Host "Extracting WiX toolset..." -ForegroundColor Gray
    Expand-Archive -Path $ZipPath -DestinationPath $TempWixDir -Force
    
    $CandlePath = Join-Path $TempWixDir "candle.exe"
    $LightPath = Join-Path $TempWixDir "light.exe"
}

Write-Host "Using WiX Candle: $CandlePath" -ForegroundColor Gray
Write-Host "Using WiX Light: $LightPath" -ForegroundColor Gray

# 3. Compile wxs to wixobj
Write-Host "Running Candle compiler..." -ForegroundColor Cyan
& $CandlePath -nologo piplex.wxs -out piplex.wixobj
if ($LASTEXITCODE -ne 0) {
    Write-Error "Candle compilation failed!"
    exit 1
}

# 4. Link wixobj to MSI
Write-Host "Linking MSI installer..." -ForegroundColor Cyan
& $LightPath -nologo piplex.wixobj -out piplex.msi
if ($LASTEXITCODE -ne 0) {
    Write-Error "Linking failed!"
    exit 1
}

# Clean up build artifacts
if (Test-Path piplex.wixobj) { Remove-Item piplex.wixobj }
if (Test-Path piplex.wixpdb) { Remove-Item piplex.wixpdb }

Write-Host "------------------------------------------------" -ForegroundColor Green
Write-Host "Success! Generated installer: $ProjectDir\piplex.msi" -ForegroundColor Green
Write-Host "------------------------------------------------" -ForegroundColor Green
