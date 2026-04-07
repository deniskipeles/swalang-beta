#!/usr/bin/env bash
# build-zlib-windows.sh  –  cross-compile zlib for Windows x64 (MinGW-w64)
set -euo pipefail

# ------------------------------------------------------------------
# 0.  Setup Temporary CMake
# ------------------------------------------------------------------
TEMP_CMAKE_VERSION="3.28.3"
TEMP_CMAKE_ROOT="$HOME/.local/cmake-${TEMP_CMAKE_VERSION}"

if [[ ! -x "$TEMP_CMAKE_ROOT/bin/cmake" ]]; then
  echo "⚙️  Fetching temporary CMake ${TEMP_CMAKE_VERSION} …"
  mkdir -p "$TEMP_CMAKE_ROOT"
  curl -sL "https://github.com/Kitware/CMake/releases/download/v${TEMP_CMAKE_VERSION}/cmake-${TEMP_CMAKE_VERSION}-linux-x86_64.tar.gz" \
    | tar -xz -C "$TEMP_CMAKE_ROOT" --strip-components=1
fi
export PATH="$TEMP_CMAKE_ROOT/bin:$PATH"
echo "Using $(cmake --version | head -1)"

# ------------------------------------------------------------------
echo "📦 1. Install MinGW-w64 toolchain if missing"
sudo apt-get update -qq && sudo apt-get install -y \
    build-essential git mingw-w64 ninja-build

echo "🧬 2. Clone zlib 1.3.1 tag"
ZLIB_VERSION="v1.3.1"
LIBRARY_NAME="zlib"
REPO_DIR=".extensions/$LIBRARY_NAME"
if [[ ! -d "$REPO_DIR" ]]; then
    git clone --depth 1 --branch ${ZLIB_VERSION} https://github.com/madler/zlib.git  "$REPO_DIR"
fi
cd "$REPO_DIR"

echo "🔨 3. Cross-compile for Windows (x86_64) using CMake"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
rm -rf build # Clean stale cache

cmake -B build \
  -G Ninja \
  -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON

cmake --build build -j"$(nproc)"
cmake --install build

echo "🧾 4. Show results"
tree "$INSTALL_DIR"