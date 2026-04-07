#!/usr/bin/env bash
set -euo pipefail

echo "📦 1. Install MinGW-w64 + CMake"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 cmake ninja-build

echo "🧬 2. Clone official TinyXML-2"
REPO_DIR="tinyxml2"
[[ -d "$REPO_DIR" ]] || \
    git clone --depth=1 https://github.com/leethomason/tinyxml2.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake cross-compile"
INSTALL_DIR="$(pwd)/../precompiled-binaries/windows-x86_64"
cmake -B build \
  -G Ninja \
  -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DBUILD_TESTING=OFF

cmake --build build -j"$(nproc)"
cmake --install build

echo "🧾 4. Results"
tree "$INSTALL_DIR"