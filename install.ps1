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

# 2. Get target version tag (queries GitHub API dynamically)
$Repo = "pphlx/pphlx"
$Tag = "latest"
$DownloadUrl = ""

if ($Tag -eq "latest") {
    try {
        $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        $Tag = $ReleaseInfo.tag_name
        
        # Look for matching Windows asset in release
        $WinAsset = $ReleaseInfo.assets | Where-Object { $_.name -match "windows" -or $_.name -match "win32" -or $_.name -match "win64" } | Select-Object -First 1
        if ($WinAsset) {
            $DownloadUrl = $WinAsset.browser_download_url
        }
    }
    catch {
        # Fallback if API fails/rate limited
        $Tag = "v1.1.6"
    }
}

if (-not $DownloadUrl) {
    $ZipName = "pphlx-windows-$Arch.zip"
    $DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/$ZipName"
}

Write-Host "Detected System: Windows ($Arch)" -ForegroundColor Green
Write-Host "Target Release:  $Tag" -ForegroundColor Green
Write-Host "Downloading:     $DownloadUrl" -ForegroundColor Gray

# 3. Download to Temp Directory
$TempDir = Join-Path $env:TEMP "pphlx-installer-$(Get-Random)"
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
$ZipPath = Join-Path $TempDir "pphlx.zip"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing
}
catch {
    Write-Host "Note: Release asset file ($DownloadUrl) is not uploaded yet." -ForegroundColor Yellow
    Write-Host "You can run PPHLX locally or via npm: npm create pphlx@latest" -ForegroundColor Cyan
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    exit 0
}

# 4. Extract Zip
Write-Host "Extracting files..." -ForegroundColor Gray
$InstallDir = Join-Path $HOME ".pphlx\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force
Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue

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
Write-Host "  Verify by running: pphlx --version" -ForegroundColor Green
Write-Host "------------------------------------------------" -ForegroundColor Cyan
