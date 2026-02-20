#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

# Version info from Git
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

LDFLAGS="-s -w"
LDFLAGS+=" -X github.com/tafli/CallingParents/internal/version.Version=${VERSION}"
LDFLAGS+=" -X github.com/tafli/CallingParents/internal/version.Commit=${COMMIT}"
LDFLAGS+=" -X github.com/tafli/CallingParents/internal/version.Date=${DATE}"

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
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/calling-parents-linux-amd64 ./cmd/server

echo "==> Building windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o dist/calling-parents-windows-amd64.exe ./cmd/server

echo "==> Copying distribution files..."
cp config.toml.example dist/
cp children.json.example dist/
cp run.sh dist/
cp run.bat dist/

echo "==> Creating archives..."
(cd dist/ && tar czf calling-parents-linux-amd64.tar.gz calling-parents-linux-amd64 config.toml.example children.json.example run.sh)
(cd dist/ && zip -q calling-parents-windows-amd64.zip calling-parents-windows-amd64.exe config.toml.example children.json.example run.bat)

echo "==> Generating checksums..."
(cd dist/ && sha256sum calling-parents-linux-amd64.tar.gz calling-parents-windows-amd64.zip > checksums.txt)

echo ""
echo "==> Build complete:"
ls -lh dist/
echo ""
echo "==> Checksums:"
cat dist/checksums.txt
