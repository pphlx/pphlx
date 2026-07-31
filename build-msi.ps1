# PPHLX MSI Installer Auto-Builder Script
# This script downloads WiX Toolset binaries temporarily and compiles pphlx.msi into releases/

$ProjectDir = Get-Location
$ReleasesDir = Join-Path $ProjectDir "releases"
$TempWixDir = Join-Path $ProjectDir ".wix-bin"

if (!(Test-Path $ReleasesDir)) {
    New-Item -ItemType Directory -Force -Path $ReleasesDir | Out-Null
}

# 1. Compile the Windows Go binary first
Write-Host "Compiling Windows Go binary for MSI..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o pphlx.exe .
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
& $CandlePath -nologo pphlx.wxs -out pphlx.wixobj
if ($LASTEXITCODE -ne 0) {
    Write-Error "Candle compilation failed!"
    exit 1
}

# 4. Link wixobj to MSI inside releases/
$MsiOutPath = Join-Path $ReleasesDir "pphlx.msi"
Write-Host "Linking MSI installer to $MsiOutPath..." -ForegroundColor Cyan
& $LightPath -nologo pphlx.wixobj -out $MsiOutPath
if ($LASTEXITCODE -ne 0) {
    Write-Error "Linking failed!"
    exit 1
}

# Clean up temporary build artifacts
if (Test-Path pphlx.wixobj) { Remove-Item pphlx.wixobj -Force }
if (Test-Path pphlx.wixpdb) { Remove-Item pphlx.wixpdb -Force }
if (Test-Path pphlx.exe) { Remove-Item pphlx.exe -Force }

Write-Host "------------------------------------------------" -ForegroundColor Green
Write-Host "Success! Generated MSI installer: $MsiOutPath" -ForegroundColor Green
Write-Host "------------------------------------------------" -ForegroundColor Green
