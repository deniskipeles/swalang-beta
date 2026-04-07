#!/usr/bin/env bash
# build-oniguruma-windows.sh  –  cross-compile Oniguruma for Windows x64 (MinGW-w64)
set -euo pipefail

echo "📦 1. Install MinGW-w64 + autotools"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 autotools-dev

echo "🧬 2. Clone official Oniguruma"
LIBRARY_NAME="oniguruma"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || \
    git clone --depth=1 https://github.com/kkos/oniguruma.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Autotools bootstrap"
autoreconf -vfi

echo "⚙️  4. Configure for Windows x64 cross-compile"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
./configure \
  --host=x86_64-w64-mingw32 \
  --prefix="$INSTALL_DIR" \
  --enable-shared \
  --enable-static \
  --disable-posix-api

echo "⚙️  5. Build & install"
make -j"$(nproc)"
make install

echo "🧾 6. Results"
tree "$INSTALL_DIR"