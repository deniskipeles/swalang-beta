#!/usr/bin/env bash
# build-civetweb-linux.sh – build Civetweb for native Linux x86-64
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
echo "📦 1. Install build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential ninja-build libssl-dev

# ------------------------------------------------------------------
LIB_NAME="civetweb"
LIB_DIR=".extensions/$LIB_NAME"
INSTALL_DIR="$(pwd)/bin/linux-x86_64/$LIB_NAME"

echo "🔄 2. Synchronising Civetweb submodule"
git submodule update --init --recursive --depth=1 -- "$LIB_DIR"
cd "$LIB_DIR"

# ------------------------------------------------------------------
echo "🔨 3. CMake native build for Linux x64"
# Clean previous build artifacts for a fresh start
rm -rf build

# Define the installation directory for CMake
cmake -B build \
      -G Ninja \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
      -DBUILD_SHARED_LIBS=ON \
      -DCIVETWEB_ENABLE_SSL=ON \
      -DCIVETWEB_BUILD_EXECUTABLE=OFF \
      -DCIVETWEB_ENABLE_WEBSOCKETS=ON \
      -DCIVETWEB_BUILD_TESTING=OFF

# ------------------------------------------------------------------
echo "⚙️  4. Build and install"
# Let CMake handle the installation process
cmake --build build -j"$(nproc)"
cmake --install build

# ------------------------------------------------------------------
echo "🧾 5. Resulting file structure"
tree "$INSTALL_DIR"

echo "✅ Build finished for $LIB_NAME."