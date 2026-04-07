#!/usr/bin/env bash
# build-libpng-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build zlib1g-dev

echo "📥 2. Clone libpng"
LIBRARY_NAME="libpng"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/glennrp/libpng.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
cmake -B build \
  -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DPNG_SHARED=ON \
  -DPNG_STATIC=OFF
cmake --build build -j"$(nproc)"
cmake --install build

tree "$INSTALL_DIR"