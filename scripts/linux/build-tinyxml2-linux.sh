#!/usr/bin/env bash
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build

echo "📥 2. Clone TinyXML-2"
REPO_DIR="tinyxml2"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/leethomason/tinyxml2.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared)"
INSTALL_DIR="$(pwd)/../precompiled-binaries/linux-x86_64"
cmake -B build \
  -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DBUILD_TESTING=OFF
cmake --build build -j"$(nproc)"
cmake --install build

tree "$INSTALL_DIR"