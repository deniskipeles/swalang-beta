#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

OUTPUT_BASE_NAME="swalang"
PACKAGE_PATH="cmd/interpreter/main.go"
BUILD_DIR="builds"

mkdir -p "$BUILD_DIR"

# Build for Linux
echo "🔨 Building interpreter for Linux x86_64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o "$BUILD_DIR/linux-x86_64/$OUTPUT_BASE_NAME" "$PACKAGE_PATH"
# Copy shared libs
mkdir -p "$BUILD_DIR/linux-x86_64/bin/x86_64-linux"
cp -r bin/x86_64-linux/* "$BUILD_DIR/linux-x86_64/bin/x86_64-linux/"

# Build for Windows
echo "🔨 Building interpreter for Windows x86_64..."
# Using zig as the linker for cross-compiling Go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="zig cc -target x86_64-windows-gnu" go build -ldflags="-s -w" -o "$BUILD_DIR/win64-x86_64/${OUTPUT_BASE_NAME}.exe" "$PACKAGE_PATH"
# Copy shared libs
mkdir -p "$BUILD_DIR/win64-x86_64/bin/x86_64-windows-gnu"
cp -r bin/x86_64-windows-gnu/* "$BUILD_DIR/win64-x86_64/bin/x86_64-windows-gnu/"

echo "✅ Interpreter built successfully!"
