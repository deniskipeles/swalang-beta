#!/usr/bin/env bash
# build-libarchive-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y \
  build-essential git cmake ninja-build \
  zlib1g-dev libbz2-dev liblzma-dev liblz4-dev libzstd-dev libxml2-dev

echo "📥 2. Clone libarchive (with sub-modules)"
LIBRARY_NAME="libarchive"
REPO_DIR=".extensions/$LIBRARY_NAME"
if [[ ! -d "$REPO_DIR" ]]; then
  git clone https://github.com/libarchive/libarchive.git "$REPO_DIR"
fi
cd "$REPO_DIR"
git submodule update --init --recursive   # <-- ensure helpers are present

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/libarchive"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DENABLE_SHARED=ON \
  -DENABLE_STATIC=OFF \
  -DENABLE_TESTING=OFF

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"