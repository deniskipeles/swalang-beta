#!/usr/bin/env bash
# build-libgit2-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git cmake ninja-build \
                        zlib1g-dev libssl-dev libssh2-1-dev

echo "📥 2. Clone libgit2"
LIBRARY_NAME="libgit2"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/libgit2/libgit2.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake build (shared only)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DBUILD_TESTS=OFF \
  -DUSE_SSH=ON \
  -DUSE_HTTPS=ON

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"