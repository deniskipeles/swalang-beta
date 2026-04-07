#!/usr/bin/env bash
# build-openssl-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git perl

echo "📥 2. Clone OpenSSL"
LIBRARY_NAME="openssl"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/openssl/openssl.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Configure & build"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
./Configure linux-x86_64 \
    --prefix="$INSTALL_DIR" \
    shared no-tests
make -j"$(nproc)"
make install

tree "$INSTALL_DIR"