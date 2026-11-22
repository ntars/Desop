#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export GOCACHE="${GOCACHE:-$(pwd)/.cache/go-build}"
export GOPATH="${GOPATH:-$(pwd)/.cache/gopath}"
mkdir -p "$GOCACHE" "$GOPATH"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
OUT="${OUT:-release/desop-${VERSION}}"

echo "Building Desop ${VERSION} for Windows (amd64)"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o "${OUT}/desop.exe" ./cmd/desop

cp README.md README_EN.md LICENSE LICENSE-MIT NOTICE LICENSING.md "$OUT/"
cp docs/INSTALL.md docs/ACCEPTANCE.md "$OUT/"
cp scripts/install.ps1 scripts/uninstall.ps1 scripts/acceptance.ps1 "$OUT/"
printf '%s\n' "$VERSION" > "$OUT/VERSION"
echo "Output: $OUT"

