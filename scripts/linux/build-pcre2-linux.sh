#!/usr/bin/env bash
# build-pcre2-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build

echo "📥 2. Clone PCRE2"
LIBRARY_NAME="pcre2"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --branch pcre2-10.42 --depth=1 https://github.com/PCRE2Project/pcre2.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DPCRE2_BUILD_TESTS=OFF

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"