#!/bin/bash
set -euo pipefail

# Create a central deps folder
mkdir -p .deps
cd .deps

if [ ! -d "libffi-3.4.6" ]; then
    echo "📥 Downloading libffi source..."
    curl -LO https://github.com/libffi/libffi/releases/download/v3.4.6/libffi-3.4.6.tar.gz
    tar -xf libffi-3.4.6.tar.gz
fi

# --- 1. Windows (x86_64-windows-gnu) ---
echo "🔨 Building static libffi for Windows..."
mkdir -p out_windows
cd libffi-3.4.6
make clean || true
CC="zig cc -target x86_64-windows-gnu" \
CXX="zig c++ -target x86_64-windows-gnu" \
./configure --host=x86_64-w64-mingw32 --enable-static --disable-shared --prefix="$(pwd)/../out_windows"
make -j$(nproc)
make install
cd ..

# --- 2. Linux Musl (x86_64-linux-musl) ---
echo "🔨 Building static libffi for Linux (Musl)..."
mkdir -p out_linux_musl
cd libffi-3.4.6
make clean || true
CC="zig cc -target x86_64-linux-musl" \
CXX="zig c++ -target x86_64-linux-musl" \
./configure --host=x86_64-linux-musl --enable-static --disable-shared --prefix="$(pwd)/../out_linux_musl"
make -j$(nproc)
make install
cd ..

echo "✅ libffi built successfully for Windows and Linux-musl."