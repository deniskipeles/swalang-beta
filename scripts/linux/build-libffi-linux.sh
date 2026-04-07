#!/usr/bin/env bash
# build-libffi-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git autotools-dev

echo "📥 2. Clone libffi"
LIBRARY_NAME="libffi"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/libffi/libffi.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Autotools bootstrap"
autoreconf -fi

echo "⚙️  4. Configure & build"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
./configure --prefix="$INSTALL_DIR" --enable-shared --disable-static
make -j"$(nproc)"
make install

tree "$INSTALL_DIR"