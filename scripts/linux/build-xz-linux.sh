#!/usr/bin/env bash
# build-xz-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git autotools-dev autopoint

echo "📥 2. Clone xz"
LIBRARY_NAME="xz"
REPO_DIR=".extensions/$LIBRARY_NAME"
# [[ -d "$REPO_DIR" ]] || git clone --depth=1 https://git.tukaani.org/xz.git "$REPO_DIR"
[[ -d "$REPO_DIR" ]] || git submodule add  --depth=1 https://git.tukaani.org/xz.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Autotools bootstrap & configure"
autoreconf -fi
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
./configure --prefix="$INSTALL_DIR" --enable-shared --disable-static
make -j"$(nproc)"
make install

tree "$INSTALL_DIR"