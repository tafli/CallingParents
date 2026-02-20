#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

# Build first
echo "==> Building for current platform..."
gofmt -w .
go vet ./...
go test ./...
go build -o calling-parents ./cmd/server

echo "==> Starting server..."
exec ./calling-parents
