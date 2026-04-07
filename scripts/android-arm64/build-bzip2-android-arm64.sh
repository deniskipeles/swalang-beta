#!/usr/bin/env bash
# build-bzip2-android-arm64.sh
set -euo pipefail

NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
HOST_TAG="linux-x86_64"
TOOLCHAIN="$NDK/toolchains/llvm/prebuilt/$HOST_TAG"
export PATH="$TOOLCHAIN/bin:$PATH"
API=24
TRIPLE="aarch64-linux-android"
CC="${TOOLCHAIN}/bin/${TRIPLE}${API}-clang"
LD="${TOOLCHAIN}/bin/${TRIPLE}${API}-clang"   # use clang for link as well

LIBRARY_NAME="bzip2"
REPO_DIR=".extensions/$LIBRARY_NAME"
mkdir -p "$REPO_DIR"

BZIP2_VER=1.0.8
curl -L "https://sourceware.org/pub/bzip2/bzip2-${BZIP2_VER}.tar.gz" -o "/tmp/bzip2-${BZIP2_VER}.tar.gz"
tar -xzf "/tmp/bzip2-${BZIP2_VER}.tar.gz" -C "$REPO_DIR" --strip-components=1
rm "/tmp/bzip2-${BZIP2_VER}.tar.gz"
cd "$REPO_DIR"

INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
mkdir -p "$INSTALL_DIR"

# ------------------------------------------------------------------
# 1.  cross-compile objects (shared & static)
# ------------------------------------------------------------------
make clean
make -j"$(nproc)" \
     CC="$CC" \
     AR="${TOOLCHAIN}/bin/llvm-ar" \
     RANLIB="${TOOLCHAIN}/bin/llvm-ranlib" \
     CFLAGS="-O2 -fPIC" \
     libbz2.a

# ------------------------------------------------------------------
# 2.  cross-compile shared library
# ------------------------------------------------------------------
make -f Makefile-libbz2_so \
     CC="$CC" \
     LD="$LD" \
     CFLAGS="-O2 -fPIC" \
     -j"$(nproc)"

# ------------------------------------------------------------------
# 3.  install
# ------------------------------------------------------------------
cp -a libbz2.so* "$INSTALL_DIR/"
cp -a bzlib.h    "$INSTALL_DIR/"

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"


echo "Build finished for $REPO_DIR."

echo "Removing source files"

# remove the bzip2 source code after building it
rm -rf "$REPO_DIR"