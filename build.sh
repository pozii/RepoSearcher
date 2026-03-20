#!/bin/bash
set -e

VERSION=${1:-"v1.0.0"}
OUTPUT_DIR="build"

mkdir -p "$OUTPUT_DIR"

echo "Building repo-searcher $VERSION..."

# Windows AMD64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -o "$OUTPUT_DIR/repo-searcher-windows-amd64.exe" .

# Windows ARM64
echo "Building for Windows ARM64..."
GOOS=windows GOARCH=arm64 go build -o "$OUTPUT_DIR/repo-searcher-windows-arm64.exe" .

# macOS AMD64 (Intel)
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -o "$OUTPUT_DIR/repo-searcher-darwin-amd64" .

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o "$OUTPUT_DIR/repo-searcher-darwin-arm64" .

# Linux AMD64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/repo-searcher-linux-amd64" .

# Linux ARM64
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/repo-searcher-linux-arm64" .

echo ""
echo "Build complete! Binaries in $OUTPUT_DIR/:"
ls -la "$OUTPUT_DIR/"
