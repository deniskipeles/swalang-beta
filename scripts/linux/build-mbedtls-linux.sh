#!/usr/bin/env bash
# build-mbedtls-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build

echo "📥 2. Clone mbed-TLS"
LIBRARY_NAME="mbedtls"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --recurse-submodules --depth=1 https://github.com/Mbed-TLS/mbedtls.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DUSE_SHARED_MBEDTLS_LIBRARY=ON \
  -DENABLE_TESTING=OFF \
  -DCMAKE_INSTALL_RPATH='$ORIGIN'

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR/lib"