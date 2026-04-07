#!/usr/bin/env bash
# build-cjson-windows.sh  –  cross-compile cJSON for Windows x64 (MinGW-w64)

set -euo pipefail

LIB_NAME="cjson"
EXT_DIR=".extensions/${LIB_NAME}"
PREFIX="$(pwd)/bin/windows-x86_64/${LIB_NAME}"

# ------------------------------------------------------------------
echo "📦 1. Install MinGW-w64 + CMake"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 cmake ninja-build tree

# ------------------------------------------------------------------
echo "🧬 2. Ensure GitHub submodule is present"
git submodule update --init --recursive --depth=1 -- "${EXT_DIR}"
cd "${EXT_DIR}"

# ------------------------------------------------------------------
echo "🔨 3. CMake cross-compile (in-source, no build dir)"
rm -rf CMakeCache.txt CMakeFiles/  # purge any stale cache
mkdir -p "${PREFIX}"

cmake -S . -B . -G Ninja \
  -DCMAKE_SYSTEM_NAME=Windows \
  -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="${PREFIX}" \
  -DBUILD_SHARED_LIBS=ON \
  -DENABLE_CJSON_TEST=OFF \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5

cmake --build . -j"$(nproc)"
cmake --install .

# ------------------------------------------------------------------
echo "🧾 4. Results"
tree "${PREFIX}"