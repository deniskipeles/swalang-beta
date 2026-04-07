#!/usr/bin/env bash
# build-civetweb-windows.sh – cross-compile Civetweb for Windows x64 (MinGW-w64)
set -euo pipefail

# ------------------------------------------------------------------
# 0.  Setup Temporary CMake to ensure a modern version is used
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
echo "✅ Using $(cmake --version | head -1)"

# ------------------------------------------------------------------
echo "📦 1. Install MinGW-w64 + base build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 ninja-build

# ------------------------------------------------------------------
echo "🔗 2. Build OpenSSL dependency first"
# make sure to uncomment this line if you are building in linux and you have not build openssl
# "$(dirname "$0")/build-openssl-windows.sh"
OPENSSL_INSTALL_DIR="$(pwd)/bin/windows-x86_64/openssl"

# ------------------------------------------------------------------
LIB_NAME="civetweb"
LIB_DIR=".extensions/$LIB_NAME"
INSTALL_DIR="$(pwd)/bin/windows-x86_64/$LIB_NAME"

echo "🔄 3. Synchronising Civetweb submodule"
git submodule update --init --recursive --depth=1 -- "$LIB_DIR"
cd "$LIB_DIR"

# ------------------------------------------------------------------
echo "🔨 4. CMake cross-compile for Windows x64"
rm -rf build # Clean previous build artifacts

# FIX: Use CMAKE_C_STANDARD_LIBRARIES to ensure libraries are appended at the end.
WINDOWS_STD_LIBS="-lws2_32 -lmswsock -lwinmm"

cmake -B build \
      -G Ninja \
      -DCMAKE_SYSTEM_NAME=Windows \
      -DCMAKE_C_COMPILER=x86_64-w64-mingw32-gcc \
      -DCMAKE_CXX_COMPILER=x86_64-w64-mingw32-g++ \
      -DCMAKE_C_STANDARD_LIBRARIES="${WINDOWS_STD_LIBS}" \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
      -DCMAKE_PREFIX_PATH="$OPENSSL_INSTALL_DIR" \
      -DBUILD_SHARED_LIBS=ON \
      -DCIVETWEB_ENABLE_SSL=ON \
      -DCIVETWEB_BUILD_EXECUTABLE=OFF \
      -DCIVETWEB_ENABLE_WEBSOCKETS=ON \
      -DCIVETWEB_BUILD_TESTING=OFF

# ------------------------------------------------------------------
echo "⚙️  5. Build and install"
cmake --build build -j"$(nproc)"
cmake --install build

# ------------------------------------------------------------------
echo "🧾 6. Resulting file structure"
tree "$INSTALL_DIR"

echo "✅ Build finished for $LIB_NAME."