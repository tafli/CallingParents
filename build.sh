#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

# Version info from Git
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

LDFLAGS="-X github.com/tafli/calling-parents/internal/version.Version=${VERSION}"
LDFLAGS+=" -X github.com/tafli/calling-parents/internal/version.Commit=${COMMIT}"
LDFLAGS+=" -X github.com/tafli/calling-parents/internal/version.Date=${DATE}"

echo "==> Version: ${VERSION} (${COMMIT}) ${DATE}"
echo ""

echo "==> Formatting..."
gofmt -w .

echo "==> Vetting..."
go vet ./...

echo "==> Testing..."
go test ./...

echo "==> Cleaning dist/..."
rm -rf dist/
mkdir -p dist/

echo "==> Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/calling_parents-linux-amd64 ./cmd/server

echo "==> Building windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/calling_parents-windows-amd64.exe ./cmd/server

echo "==> Copying distribution files..."
cp config.toml.example dist/
cp children.json.example dist/
cp run.bat dist/

echo ""
echo "==> Build complete:"
ls -lh dist/
