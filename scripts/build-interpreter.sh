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
if [ ! -d "$PROJECT_ROOT/.deps/out_windows" ] ||[ ! -d "$PROJECT_ROOT/.deps/out_linux_gnu" ]; then
    echo "⚠️  Static libffi dependencies not found or outdated. Running cross-compile script..."
    bash scripts/setup/cross-compile-libffi.sh
fi

# ---------------------------------------------------------
# 1. Build for Linux (Max Compatibility via glibc 2.17)
# ---------------------------------------------------------
echo "🔨 Building interpreter for Linux x86_64 (glibc 2.17)..."
LINUX_DIR="$BUILD_DIR/linux-x86_64"
mkdir -p "$LINUX_DIR/bin" "$LINUX_DIR/lib" "$LINUX_DIR/stdlib"

# By targeting an ancient glibc, this binary runs on virtually ALL Linux machines while keeping dlopen() support!
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
CC="zig cc -target x86_64-linux-gnu.2.17" \
CXX="zig c++ -target x86_64-linux-gnu.2.17" \
CGO_CFLAGS="-I$PROJECT_ROOT/.deps/out_linux_gnu/include" \
CGO_LDFLAGS="-L$PROJECT_ROOT/.deps/out_linux_gnu/lib" \
go build -tags="$TAGS" -ldflags="-s -w" -o "$LINUX_DIR/bin/$OUTPUT_BASE_NAME" "$PACKAGE_PATH"

# Copy shared libs & stdlib into production structure
cp -r bin/x86_64-linux/* "$LINUX_DIR/lib/" 2>/dev/null || true
cp -r stdlib/* "$LINUX_DIR/stdlib/" 2>/dev/null || true

# ---------------------------------------------------------
# 2. Build for Windows
# ---------------------------------------------------------
echo "🔨 Building interpreter for Windows x86_64..."
WIN_DIR="$BUILD_DIR/win64-x86_64"
mkdir -p "$WIN_DIR/bin" "$WIN_DIR/lib" "$WIN_DIR/stdlib"

GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
CC="zig cc -target x86_64-windows-gnu" \
CXX="zig c++ -target x86_64-windows-gnu" \
CGO_CFLAGS="-I$PROJECT_ROOT/.deps/out_windows/include" \
CGO_LDFLAGS="-L$PROJECT_ROOT/.deps/out_windows/lib" \
go build -tags="$TAGS" -ldflags="-s -w" -o "$WIN_DIR/bin/${OUTPUT_BASE_NAME}.exe" "$PACKAGE_PATH"

# Copy shared libs & stdlib into production structure
cp -r bin/x86_64-windows-gnu/* "$WIN_DIR/lib/" 2>/dev/null || true
cp -r stdlib/* "$WIN_DIR/stdlib/" 2>/dev/null || true

echo "✅ Interpreter built and packaged successfully!"