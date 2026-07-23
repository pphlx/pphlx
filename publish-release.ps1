$ErrorActionPreference = "Stop"

$ProjectDir = Get-Location
$ReleasesDir = Join-Path $ProjectDir "releases"
$Version = "v1.1.0"
$ReleaseTitle = "PPHLX Compiler $Version"
$NotesFile = Join-Path $ReleasesDir "release-notes-v1.1.0.md"

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host " Publishing $ReleaseTitle" -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# 1. Build all 5 release archives
Write-Host "`n[Step 1/4] Building cross-platform release binaries..." -ForegroundColor Yellow
& "$ProjectDir\build-release.ps1"

# 2. Tag current commit
Write-Host "`n[Step 2/4] Tagging git commit with $Version..." -ForegroundColor Yellow
$ExistingTag = git tag -l $Version
if (-not $ExistingTag) {
    git tag $Version
    Write-Host "Created git tag: $Version" -ForegroundColor Green
} else {
    Write-Host "Git tag $Version already exists." -ForegroundColor Gray
}

# 3. Push commit & tags to remote
Write-Host "`n[Step 3/4] Pushing main branch and tags to GitHub..." -ForegroundColor Yellow
git push -u origin main --tags

# 4. Publish GitHub Release using gh CLI
Write-Host "`n[Step 4/4] Creating GitHub Release using gh CLI..." -ForegroundColor Yellow

$Assets = @(
    (Join-Path $ReleasesDir "pphlx-darwin-arm64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-darwin-amd64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-linux-arm64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-linux-amd64.tar.gz"),
    (Join-Path $ReleasesDir "pphlx-windows-amd64.zip")
)

# Verify assets exist
foreach ($Asset in $Assets) {
    if (-not (Test-Path $Asset)) {
        Write-Error "Missing required release asset: $Asset"
    }
}

# Run gh release create
gh release create $Version $Assets --title $ReleaseTitle --notes-file $NotesFile --repo pphlx/pphlx

Write-Host "`n=================================================" -ForegroundColor Green
Write-Host " Successfully published $ReleaseTitle to GitHub!" -ForegroundColor Green
Write-Host " Release URL: https://github.com/pphlx/pphlx/releases/tag/$Version" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
