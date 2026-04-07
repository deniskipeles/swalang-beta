#!/usr/bin/env bash
# build-libjpeg-turbo-windows.sh  –  cross-compile libjpeg-turbo for Windows x64 (MinGW-w64)
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
echo "📦 1. Install MinGW-w64 + build tools"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 ninja-build

echo "🧬 2. Clone official libjpeg-turbo (specific version)"
LIBRARY_NAME="libjpeg-turbo"
REPO_DIR=".extensions/$LIBRARY_NAME"
# Reverting to the version from your logs
[[ -d "$REPO_DIR" ]] || \
    git clone --branch 2.1.5.1 --depth=1 https://github.com/libjpeg-turbo/libjpeg-turbo.git  "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. CMake cross-compile (Windows x64)"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
rm -rf build # Clean stale cache

cmake -B build \
  -G Ninja \
  -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_SYSTEM_PROCESSOR=x86_64 \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DENABLE_SHARED=ON \
  -DENABLE_STATIC=OFF \
  -DWITH_JPEG8=ON

echo "⚙️  4. Build & install"
cmake --build build -j"$(nproc)"
cmake --install build

echo "🧾 5. Results"
tree "$INSTALL_DIR"