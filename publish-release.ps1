$ErrorActionPreference = "Stop"

$CoreDir = Get-Location
$RootDir = Split-Path $CoreDir -Parent
$ReleasesDir = Join-Path $CoreDir "releases"

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
$TagVersion = "v$Version"
$ReleaseTitle = "PPHLX Compiler $TagVersion"
$NotesFile = Join-Path $ReleasesDir "release-notes-v$Version.md"

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " Publishing $ReleaseTitle to GitHub Releases" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# 2. Check if GitHub release notes exist
if (-not (Test-Path $NotesFile)) {
    Write-Host "Warning: Release notes file not found at $NotesFile. Creating template..." -ForegroundColor Yellow
    "## PPHLX Compiler $TagVersion Release Notes" | Out-File -FilePath $NotesFile -Encoding utf8
}

# 3. Publish GitHub Release using gh CLI
$Assets = @(
    (Join-Path $ReleasesDir "pphlx-darwin-arm64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-darwin-amd64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-linux-arm64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-linux-amd64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-windows-amd64.zip"),
    (Join-Path $ReleasesDir "pphlx.msi")
)

# Verify assets exist
foreach ($Asset in $Assets) {
    if (-not (Test-Path $Asset)) {
        Write-Error "Missing required release asset: $Asset. Please run build-release.ps1 first."
        exit 1
    }
}

Write-Host "Publishing GitHub release for tag $TagVersion..." -ForegroundColor Yellow
gh release create $TagVersion $Assets --title $ReleaseTitle --notes-file $NotesFile --repo pphlx/pphlx

Write-Host "`n=================================================" -ForegroundColor Green
Write-Host " Successfully published $ReleaseTitle to GitHub!" -ForegroundColor Green
Write-Host " Release URL: https://github.com/pphlx/pphlx/releases/tag/$TagVersion" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
