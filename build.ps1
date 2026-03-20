$ErrorActionPreference = "Stop"

$VERSION = if ($args[0]) { $args[0] } else { "v1.0.0" }
$OUTPUT_DIR = "build"
$COMMIT = try { git rev-parse --short HEAD } catch { "unknown" }
$BUILD_DATE = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LDFLAGS = "-s -w -X github.com/pozii/RepoSearcher/cmd.Version=$VERSION -X github.com/pozii/RepoSearcher/cmd.GitCommit=$COMMIT -X github.com/pozii/RepoSearcher/cmd.BuildDate=$BUILD_DATE"

New-Item -ItemType Directory -Force -Path $OUTPUT_DIR | Out-Null

Write-Host "Building repo-searcher $VERSION (commit: $COMMIT)..." -ForegroundColor Cyan
Write-Host "Flags: CGO_ENABLED=0 -trimpath -ldflags" -ForegroundColor DarkGray

function Build-Target {
    param($OS, $Arch, $Suffix)
    Write-Host "Building for $OS/$Arch..." -ForegroundColor Yellow
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    $env:CGO_ENABLED = "0"
    go build -trimpath -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/repo-searcher-${OS}-${Arch}${Suffix}" .
}

Build-Target "windows" "amd64" ".exe"
Build-Target "windows" "arm64" ".exe"
Build-Target "darwin" "amd64" ""
Build-Target "darwin" "arm64" ""
Build-Target "linux" "amd64" ""
Build-Target "linux" "arm64" ""

# Reset environment
$env:GOOS = $null
$env:GOARCH = $null
$env:CGO_ENABLED = $null

# Generate checksums
Write-Host "Generating checksums..." -ForegroundColor Yellow
$files = Get-ChildItem -Path $OUTPUT_DIR -Filter "repo-searcher-*"
$checksumLines = @()
foreach ($f in $files) {
    $hash = (Get-FileHash -Path $f.FullName -Algorithm SHA256).Hash.ToLower()
    $checksumLines += "$hash  $($f.Name)"
}
$checksumLines | Out-File -FilePath "$OUTPUT_DIR/checksums.txt" -Encoding ascii
Write-Host "checksums.txt created" -ForegroundColor Green

Write-Host ""
Write-Host "Build complete! Binaries in $OUTPUT_DIR/:" -ForegroundColor Green
Get-ChildItem -Path $OUTPUT_DIR -File | Format-Table Name, Length -AutoSize
