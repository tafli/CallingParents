#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "==> Formatting..."
gofmt -w .

echo "==> Vetting..."
go vet ./...

echo "==> Testing..."
go test ./...

echo "==> Building linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o dist/calling_parents-linux-amd64 ./cmd/server

echo "==> Building windows/amd64..."
GOOS=windows GOARCH=amd64 go build -o dist/calling_parents-windows-amd64.exe ./cmd/server

echo ""
echo "==> Build complete:"
ls -lh dist/
