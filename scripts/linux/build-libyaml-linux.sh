#!/usr/bin/env bash
# build-libyaml-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake

echo "📥 2. Clone libyaml"
LIBRARY_NAME="libyaml"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/yaml/libyaml.git "$REPO_DIR"
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