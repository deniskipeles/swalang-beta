#!/usr/bin/env bash
# build-xz-windows.sh  –  cross-compile XZ for Windows x64 (MinGW-w64)
set -euo pipefail

echo "📦 1. Install MinGW-w64 toolchain"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 autotools-dev autopoint

echo "🧬 2. Clone official XZ source"
LIBRARY_NAME="xz"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://git.tukaani.org/xz.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Autotools bootstrap & configure"
autoreconf -fi
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
./configure \
  --host=x86_64-w64-mingw32 \
  --prefix="$INSTALL_DIR" \
  --enable-shared \
  --enable-static

echo "⚙️  4. Build & install"
make -j"$(nproc)"
make install

echo "🧾 5. Results"
tree "$INSTALL_DIR"