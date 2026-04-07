#!/usr/bin/env bash
# build-libjpeg-turbo-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build

echo "📥 2. Clone libjpeg-turbo"
LIBRARY_NAME="libjpeg-turbo"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --branch 2.1.5.1 --depth=1 https://github.com/libjpeg-turbo/libjpeg-turbo.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DENABLE_STATIC=OFF \
  -DWITH_JPEG8=ON \
  -DWITH_SIMD=ON \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"