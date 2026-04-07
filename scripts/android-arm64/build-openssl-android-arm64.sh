#!/usr/bin/env bash
# build-openssl-android-arm64.sh
# Builds OpenSSL 3.0.15 for Android arm64-v8a (shared + static)
# Uses same NDK r26d you already keep in $HOME/android-ndk-r26d
set -euo pipefail

# ------------------------------------------------------------------
# 1.  Android NDK (re-use or auto-download)
# ------------------------------------------------------------------
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
HOST_TAG="linux-x86_64"
TOOLCHAIN="$NDK/toolchains/llvm/prebuilt/$HOST_TAG"
export ANDROID_NDK="$NDK"          # required by OpenSSL Configure
export PATH="$TOOLCHAIN/bin:$PATH"

API=24
ARCH="arm64"
TARGET="android-${ARCH}"
TRIPLE="aarch64-linux-android"
CLANG="${TOOLCHAIN}/bin/${TRIPLE}${API}-clang"

# ------------------------------------------------------------------
# 2.  OpenSSL version & install dir
# ------------------------------------------------------------------
OPENSSL_VER="3.0.15"
INSTALL_DIR="$(pwd)/bin/android/arm64-v8a/openssl"
mkdir -p "$INSTALL_DIR"

# ------------------------------------------------------------------
# 3.  Download / unpack
# ------------------------------------------------------------------
cd /tmp   # or any temp dir
[[ -f "openssl-${OPENSSL_VER}.tar.gz" ]] ||
  wget -q --show-progress "https://github.com/openssl/openssl/archive/openssl-${OPENSSL_VER}.tar.gz"
tar -xzf "openssl-${OPENSSL_VER}.tar.gz"
cd "openssl-openssl-${OPENSSL_VER}"

# ------------------------------------------------------------------
# 4.  Configure & build
# ------------------------------------------------------------------
./Configure ${TARGET} -D__ANDROID_API__=${API} \
  --prefix="$INSTALL_DIR" shared no-tests

make -j"$(nproc)"
make install_sw DESTDIR=

# ------------------------------------------------------------------
# 5.  tidy up & show result
# ------------------------------------------------------------------
cd ..
rm -rf "openssl-openssl-${OPENSSL_VER}"
echo "✅ OpenSSL ready:"
tree "$INSTALL_DIR"