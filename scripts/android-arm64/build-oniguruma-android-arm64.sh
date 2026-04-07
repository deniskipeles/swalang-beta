#!/usr/bin/env bash
set -e
# uo pipefail

# 1. Check or install Android NDK
NDK_DIR="$HOME/android-ndk-r26d"
if [ -z "$ANDROID_NDK_HOME" ] && [ -z "$ANDROID_NDK" ]; then
    if [ ! -d "$NDK_DIR" ]; then
        echo "Android NDK not found at $NDK_DIR."
        echo "Downloading Android NDK r26d..."
        wget -q --show-progress https://dl.google.com/android/repository/android-ndk-r26d-linux.zip -O "$HOME/android-ndk-r26d-linux.zip"
        unzip -q "$HOME/android-ndk-r26d-linux.zip" -d "$HOME"
        rm "$HOME/android-ndk-r26d-linux.zip"
        echo "Android NDK r26d installed at $NDK_DIR."
    else
        echo "Android NDK already present at $NDK_DIR."
    fi
    export ANDROID_NDK="$NDK_DIR"
else
    NDK_DIR="${ANDROID_NDK_HOME:-$ANDROID_NDK}"
    echo "Using existing Android NDK at $NDK_DIR."
fi

# [[ -z "${ANDROID_NDK:-}" ]] && { echo "Set ANDROID_NDK"; exit 1; }

# 2. Install oniguruma
LIBRARY_NAME="oniguruma"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 https://github.com/kkos/oniguruma.git "$REPO_DIR"
cd "$REPO_DIR"

INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
[[ -d "$INSTALL_DIR" ]] || mkdir -p "$INSTALL_DIR"
TOOLCHAIN="$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64"
API=24
TARGET=aarch64-linux-android

export CC=$TOOLCHAIN/bin/${TARGET}${API}-clang
export AR=$TOOLCHAIN/bin/llvm-ar
export RANLIB=$TOOLCHAIN/bin/llvm-ranlib
export STRIP=$TOOLCHAIN/bin/llvm-strip

autoreconf -vfi
./configure \
  --host=aarch64-linux-android \
  --prefix="$INSTALL_DIR" \
  --enable-shared \
  --disable-static \
  --disable-posix-api
make -j$(nproc)
make install

tree "$INSTALL_DIR"

echo "Build finished for $REPO_DIR."

# Unset temp variables
unset CC
unset AR
unset RANLIB
unset STRIP
# unset CXX

echo "Environment variables unset."