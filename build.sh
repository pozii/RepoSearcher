#!/bin/bash
set -e

VERSION=${1:-"v1.0.0"}
OUTPUT_DIR="build"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-s -w -X github.com/pozii/RepoSearcher/cmd.Version=${VERSION} -X github.com/pozii/RepoSearcher/cmd.GitCommit=${COMMIT} -X 'github.com/pozii/RepoSearcher/cmd.BuildDate=${BUILD_DATE}'"

mkdir -p "$OUTPUT_DIR"

echo "Building repo-searcher $VERSION (commit: $COMMIT)..."
echo "Flags: CGO_ENABLED=0 -trimpath -ldflags"

build() {
    local os=$1 arch=$2 suffix=$3
    echo "Building for $os/$arch..."
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags="$LDFLAGS" -o "$OUTPUT_DIR/repo-searcher-${os}-${arch}${suffix}" .
}

build windows amd64 .exe
build windows arm64 .exe
build darwin amd64 ""
build darwin arm64 ""
build linux amd64 ""
build linux arm64 ""

echo "Generating checksums..."
cd "$OUTPUT_DIR"
sha256sum repo-searcher-* > checksums.txt
cat checksums.txt
cd ..

echo ""
echo "Build complete! Binaries in $OUTPUT_DIR/:"
ls -la "$OUTPUT_DIR/"
