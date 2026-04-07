#!/usr/bin/env bash
# build-jansson-windows.sh  –  cross-compile jansson for Windows x64 (MinGW-w64)
set -euo pipefail

LIB_NAME="jansson"
EXT_DIR=".extensions/${LIB_NAME}"
PREFIX="$(pwd)/bin/windows-x86_64/${LIB_NAME}"
TARGET_HOST="x86_64-w64-mingw32"

# ------------------------------------------------------------------
echo "📦 1. Install MinGW-w64 + build dependencies"
# We need autotools for this build, not cmake
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 autotools-dev libtool

# ------------------------------------------------------------------
echo "🧬 2. Clone / update jansson"
if ! [[ -d "${EXT_DIR}/.git" ]]; then
  git clone --depth 1 https://github.com/akheron/jansson.git "${EXT_DIR}"
fi
cd "${EXT_DIR}"

# ------------------------------------------------------------------
echo "🧹 3. Clean any stale objects (from previous builds)"
# This is critical to switch from a failed CMake build to a clean Autotools one
git clean -xfd

# ------------------------------------------------------------------
echo "🔨 4. Autotools cross-build for Windows x64"
# Bootstrap the configure script if it's missing
[[ -f configure ]] || autoreconf -fi

# Set environment variables to point to the MinGW cross-compiler toolchain
export CC="${TARGET_HOST}-gcc"
export CXX="${TARGET_HOST}-g++"
export AR="${TARGET_HOST}-ar"
export STRIP="${TARGET_HOST}-strip"
export RANLIB="${TARGET_HOST}-ranlib"

# Configure the build for the Windows target host
./configure --host="$TARGET_HOST" \
  --prefix="$PREFIX" \
  --disable-static \
  --enable-shared \
  --with-pic

# Build and install
make -j"$(nproc)"
make install

# ------------------------------------------------------------------
echo "🧾 5. Results"
cd ../../.. # Return to the project root
tree "${PREFIX}"

# ------------------------------------------------------------------
# Unset the environment variables for a clean shell
unset CC CXX AR STRIP RANLIB
echo "✅ Done."