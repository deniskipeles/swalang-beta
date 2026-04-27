#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

OUTPUT_BASE_NAME="swalang"
PACKAGE_PATH="cmd/interpreter/main.go"
BUILD_DIR="builds"
TAGS="en" # Set default language tags here

mkdir -p "$BUILD_DIR"

# Ensure static libffi dependencies exist for both platforms
if [ ! -d "$PROJECT_ROOT/.deps/out_windows" ] || [ ! -d "$PROJECT_ROOT/.deps/out_linux_musl" ]; then
    echo "⚠️  Static libffi dependencies not found. Running cross-compile script..."
    bash scripts/setup/cross-compile-libffi.sh
fi

# ---------------------------------------------------------
# 1. Build for Linux (Musl for complete portability)
# ---------------------------------------------------------
echo "🔨 Building interpreter for Linux x86_64 (Musl)..."
# -linkmode external -extldflags '-static' forces a 100% standalone binary without dynamic libc links
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
CC="zig cc -target x86_64-linux-musl" \
CXX="zig c++ -target x86_64-linux-musl" \
CGO_CFLAGS="-I$PROJECT_ROOT/.deps/out_linux_musl/include" \
CGO_LDFLAGS="-L$PROJECT_ROOT/.deps/out_linux_musl/lib" \
go build -tags="$TAGS" -ldflags="-linkmode external -extldflags '-static' -s -w" -o "$BUILD_DIR/linux-x86_64/$OUTPUT_BASE_NAME" "$PACKAGE_PATH"

# Copy shared libs
mkdir -p "$BUILD_DIR/linux-x86_64/bin/x86_64-linux"
cp -r bin/x86_64-linux/* "$BUILD_DIR/linux-x86_64/bin/x86_64-linux/" 2>/dev/null || true

# ---------------------------------------------------------
# 2. Build for Windows
# ---------------------------------------------------------
echo "🔨 Building interpreter for Windows x86_64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
CC="zig cc -target x86_64-windows-gnu" \
CXX="zig c++ -target x86_64-windows-gnu" \
CGO_CFLAGS="-I$PROJECT_ROOT/.deps/out_windows/include" \
CGO_LDFLAGS="-L$PROJECT_ROOT/.deps/out_windows/lib" \
go build -tags="$TAGS" -ldflags="-s -w" -o "$BUILD_DIR/win64-x86_64/${OUTPUT_BASE_NAME}.exe" "$PACKAGE_PATH"

# Copy shared libs
mkdir -p "$BUILD_DIR/win64-x86_64/bin/x86_64-windows-gnu"
cp -r bin/x86_64-windows-gnu/* "$BUILD_DIR/win64-x86_64/bin/x86_64-windows-gnu/" 2>/dev/null || true

echo "✅ Interpreter built successfully!"