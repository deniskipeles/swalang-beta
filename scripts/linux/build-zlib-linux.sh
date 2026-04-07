#!/usr/bin/env bash
# build-zlib-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake

echo "📥 2. Clone zlib"
LIBRARY_NAME="zlib"
REPO_DIR=".extensions/$LIBRARY_NAME"
# [[ -d "$REPO_DIR" ]] || git clone --depth=1 --branch v1.3.1 https://github.com/madler/zlib.git "$REPO_DIR"
[[ -d "$REPO_DIR" ]] || git submodule add --depth=1 --branch v1.3.1 https://github.com/madler/zlib.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
cmake -B build \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON
cmake --build build -j"$(nproc)"
cmake --install build

tree "$INSTALL_DIR"