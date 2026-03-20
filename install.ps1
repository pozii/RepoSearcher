$ErrorActionPreference = "Stop"

$Repo = "pozii/RepoSearcher"
$Binary = "repo-searcher"
$InstallDir = "$env:LOCALAPPDATA\RepoSearcher"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Host "Error: 32-bit systems are not supported" -ForegroundColor Red
    exit 1
}

$Asset = "${Binary}-windows-${Arch}.exe"

Write-Host "Downloading ${Binary} for Windows/${Arch}..." -ForegroundColor Cyan

# Get latest release tag
$Latest = (Invoke-RestMethod -Uri "https://api.github.com/repos/${Repo}/releases/latest").tag_name

if (-not $Latest) {
    Write-Host "Error: Could not fetch latest release" -ForegroundColor Red
    exit 1
}

Write-Host "Latest version: ${Latest}" -ForegroundColor Green

# Download binary
$DownloadUrl = "https://github.com/${Repo}/releases/download/${Latest}/${Asset}"

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$TargetPath = Join-Path $InstallDir "${Binary}.exe"
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetPath

# Add to PATH if not already
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$currentPath", "User")
    $env:PATH = "$InstallDir;$env:PATH"
    Write-Host "Added $InstallDir to PATH" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "✅ ${Binary} ${Latest} installed to ${TargetPath}" -ForegroundColor Green
Write-Host ""
Write-Host "Run '${Binary} version' to verify installation." -ForegroundColor Cyan
