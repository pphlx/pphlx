# PPHLX CLI Installer for Windows PowerShell
# Downloads and installs the native Windows binary into user profile path without requiring administrator privileges.

$ErrorActionPreference = "Stop"

Write-Host "------------------------------------------------" -ForegroundColor Cyan
Write-Host "      PPHLX Windows CLI Installer Boot Sequence" -ForegroundColor Cyan
Write-Host "------------------------------------------------" -ForegroundColor Cyan

# 1. Detect CPU Architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
switch -regex ($Arch) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default {
        Write-Error "Unsupported CPU architecture: $Arch"
        exit 1
    }
}

# 2. Get target version tag (Hardcoded for dev phase)
$Repo = "KillerTyzon/PPHLX"
$Tag = "v1.0.0"

$ZipName = "PPHLX-windows-$Arch.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$ZipName"

Write-Host "Detected System: Windows ($Arch)" -ForegroundColor Green
Write-Host "Target Release:  $Tag" -ForegroundColor Green
Write-Host "Downloading:     $DownloadUrl" -ForegroundColor Gray

# 3. Download to Temp Directory
$TempDir = Join-Path $env:TEMP "PPHLX-installer-$(Get-Random)"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
$ZipPath = Join-Path $TempDir "PPHLX.zip"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
} catch {
    Write-Error "Failed to download the PPHLX release zip. It might not be uploaded yet."
    exit 1
}

# 4. Extract Zip
Write-Host "Extracting files..." -ForegroundColor Gray
$InstallDir = Join-Path $HOME ".PPHLX\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force
Remove-Item -Path $TempDir -Recurse -Force | Out-Null

# 5. Append to User PATH environment variable
Write-Host "Updating User environment PATH..." -ForegroundColor Gray
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -split ";" -notcontains $InstallDir) {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    # Refresh local session environment variable
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host "------------------------------------------------" -ForegroundColor Cyan
Write-Host "  Success! PPHLX has been installed successfully." -ForegroundColor Green
Write-Host "  Please RESTART your terminal/VS Code for PATH updates to load." -ForegroundColor Yellow
Write-Host "  Verify by running: PPHLX --version" -ForegroundColor Green
Write-Host "------------------------------------------------" -ForegroundColor Cyan

