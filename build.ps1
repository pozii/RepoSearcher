$ErrorActionPreference = "Stop"

$VERSION = if ($args[0]) { $args[0] } else { "v1.0.0" }
$OUTPUT_DIR = "build"

New-Item -ItemType Directory -Force -Path $OUTPUT_DIR | Out-Null

Write-Host "Building repo-searcher $VERSION..." -ForegroundColor Cyan

# Windows AMD64
Write-Host "Building for Windows AMD64..."
$env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -o "$OUTPUT_DIR/repo-searcher-windows-amd64.exe" .

# Windows ARM64
Write-Host "Building for Windows ARM64..."
$env:GOARCH = "arm64"
go build -o "$OUTPUT_DIR/repo-searcher-windows-arm64.exe" .

# macOS AMD64 (Intel)
Write-Host "Building for macOS AMD64..."
$env:GOOS = "darwin"; $env:GOARCH = "amd64"
go build -o "$OUTPUT_DIR/repo-searcher-darwin-amd64" .

# macOS ARM64 (Apple Silicon)
Write-Host "Building for macOS ARM64..."
$env:GOARCH = "arm64"
go build -o "$OUTPUT_DIR/repo-searcher-darwin-arm64" .

# Linux AMD64
Write-Host "Building for Linux AMD64..."
$env:GOOS = "linux"; $env:GOARCH = "amd64"
go build -o "$OUTPUT_DIR/repo-searcher-linux-amd64" .

# Linux ARM64
Write-Host "Building for Linux ARM64..."
$env:GOARCH = "arm64"
go build -o "$OUTPUT_DIR/repo-searcher-linux-arm64" .

# Reset environment
$env:GOOS = $null
$env:GOARCH = $null

Write-Host ""
Write-Host "Build complete! Binaries in $OUTPUT_DIR/:" -ForegroundColor Green
Get-ChildItem -Path $OUTPUT_DIR -File | Format-Table Name, Length -AutoSize
