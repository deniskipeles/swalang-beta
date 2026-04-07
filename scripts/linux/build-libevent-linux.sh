#!/usr/bin/env bash
# build-libevent-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build

echo "📥 2. Clone libevent"
LIBRARY_NAME="libevent"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/libevent/libevent.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DEVENT__DISABLE_TESTS=ON \
  -DEVENT__DISABLE_SAMPLES=ON \
  -DEVENT__DISABLE_REGRESS=ON

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"