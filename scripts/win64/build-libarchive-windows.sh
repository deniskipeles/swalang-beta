#!/usr/bin/env bash
# build-libarchive-windows.sh  –  cross-compile libarchive for Windows x64 (MinGW-w64)
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
echo "📦 1. Install MinGW-w64 + Build Tools"
# The base mingw-w64 package includes the necessary dependency libs (zlib, bzip2, etc.)
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 cmake ninja-build

echo "🧬 2. Clone official libarchive"
LIBRARY_NAME="libarchive"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || \
    git clone --depth=1 https://github.com/libarchive/libarchive.git  "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake cross-compile (Windows x64)"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
MINGW_ROOT="/usr/x86_64-w64-mingw32"
rm -rf build # Clean stale cache

cmake -B build \
  -G Ninja \
  -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_SYSTEM_PROCESSOR=x86_64 \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DCMAKE_FIND_ROOT_PATH="$MINGW_ROOT" \
  -DENABLE_SHARED=ON \
  -DENABLE_STATIC=OFF \
  -DENABLE_TEST=OFF

echo "⚙️  4. Build & install"
cmake --build build -j"$(nproc)"
cmake --install build

echo "🧾 5. Results"
tree "$INSTALL_DIR"