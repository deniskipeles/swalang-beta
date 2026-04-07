#!/usr/bin/env bash
# build-jansson-linux.sh
# Native Linux x86-64 shared library
set -euo pipefail

echo "⚙️  1. Install build deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential autotools-dev autoconf automake libtool curl tar

LIBRARY_NAME="jansson"
REPO_DIR=".extensions/$LIBRARY_NAME"
mkdir -p "$REPO_DIR"

echo "📥 2. Clone / update jansson"
[[ -d "$REPO_DIR/.git" ]] || git clone --depth 1 https://github.com/akheron/jansson.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🧹 3. Clean stale artefacts"
make distclean 2>/dev/null || true
git clean -xfd

echo "🔨 4. Autotools native build"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
mkdir -p "$INSTALL_DIR"

[[ -f configure ]] || autoreconf -fi

./configure --prefix="$INSTALL_DIR" \
            --disable-static \
            --enable-shared \
            --with-pic

make -j"$(nproc)"
make install

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"

echo "✅ Done."