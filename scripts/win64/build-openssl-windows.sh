#!/usr/bin/env bash
# build-openssl-windows.sh  –  cross-compile OpenSSL for Windows x64 (MinGW-w64)
set -euo pipefail

# ----------------------------------------------------------
echo "📦 1. Install MinGW-w64 + build deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 perl

# ----------------------------------------------------------
echo "🧬 2. Clone official OpenSSL source (latest master)"
LIBRARY_NAME="openssl"
REPO_DIR=".extensions/$LIBRARY_NAME"
if [[ ! -d "$REPO_DIR" ]]; then
    git clone --depth=1 https://github.com/openssl/openssl.git  "$REPO_DIR"
fi
cd "$REPO_DIR"

# ----------------------------------------------------------
echo "🔨 3. Configure for Windows x64 cross-compile"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
# Clean previous failed configuration to be safe
make clean || true

./Configure mingw64 \
    --cross-compile-prefix=x86_64-w64-mingw32- \
    --prefix="${INSTALL_DIR}" \
    shared \
    no-asm \
    no-tests

# ----------------------------------------------------------
echo "⚙️  4. Build & install"
make -j"$(nproc)"
make install

# ----------------------------------------------------------
echo "🧾 5. Result"
tree "${INSTALL_DIR}"