#!/usr/bin/env bash
# build-curl-android-arm64.sh
# Reference: https://github.com/robertying/openssl-curl-android
set -euo pipefail

# ------------------------------------------------------------------
# 0.  Android NDK (re-use the one you already have)
# ------------------------------------------------------------------
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
export ANDROID_NDK="$NDK"
HOST_TAG="linux-x86_64"
TOOLCHAIN="$NDK/toolchains/llvm/prebuilt/$HOST_TAG"
export PATH="$TOOLCHAIN/bin:$PATH"

ANDROID_API=24
ARCH="arm64"
TARGET="${ARCH}-linux-android"
TRIPLE="aarch64-linux-android"

# ------------------------------------------------------------------
# 1.  Directories
# ------------------------------------------------------------------
WORK_DIR="$(pwd)/.curl-build-tmp"      # throw-away build folder
PREFIX="$(pwd)/bin/android/arm64-v8a/curl"   # final install dir
mkdir -p "$WORK_DIR" "$PREFIX"
cd "$WORK_DIR"

# ------------------------------------------------------------------
# 2.  OpenSSL 1.1.1w (static, no-tests)
# ------------------------------------------------------------------
OPENSSL_VER="1.1.1w"
OPENSSL_SRC="openssl-${OPENSSL_VER}"
if [[ ! -d "$OPENSSL_SRC" ]]; then
  echo "Downloading OpenSSL $OPENSSL_VER …"
  wget -q --show-progress "https://www.openssl.org/source/$OPENSSL_SRC.tar.gz"
  tar -xzf "$OPENSSL_SRC.tar.gz" && rm "$OPENSSL_SRC.tar.gz"
fi
cd "$OPENSSL_SRC"

  # NDK clang
export CC="${TOOLCHAIN}/bin/${TRIPLE}${ANDROID_API}-clang"
export AR="${TOOLCHAIN}/bin/llvm-ar"
export AS="${TOOLCHAIN}/bin/llvm-as"
export NM="${TOOLCHAIN}/bin/llvm-nm"
export STRIP="${TOOLCHAIN}/bin/llvm-strip"
export RANLIB="${TOOLCHAIN}/bin/llvm-ranlib"

./Configure android-${ARCH} -D__ANDROID_API__=$ANDROID_API \
  --prefix="$PREFIX" no-shared no-tests

make -j"$(nproc)"
make install_sw DESTDIR=
cd ..

# ------------------------------------------------------------------
# 3.  curl 8.4.0 (shared library, no tool)
# ------------------------------------------------------------------
CURL_VER="8.4.0"
CURL_SRC="curl-${CURL_VER}"
if [[ ! -d "$CURL_SRC" ]]; then
  echo "Downloading curl $CURL_VER …"
  wget -q --show-progress "https://github.com/curl/curl/releases/download/curl-${CURL_VER//./_}/$CURL_SRC.tar.gz"
  tar -xzf "$CURL_SRC.tar.gz" && rm "$CURL_SRC.tar.gz"
fi
cd "$CURL_SRC"

  # tell curl where OpenSSL lives
export PKG_CONFIG_PATH="$PREFIX/lib/pkgconfig"
./configure --host="$TARGET" \
  --prefix="$PREFIX" \
  --with-ssl="$PREFIX" \
  --disable-static --enable-shared \
  --disable-ldap --disable-ldaps \
  --without-libpsl --without-librtmp \
  --without-brotli --without-zstd \
  --with-pic

make -j"$(nproc)"
make install DESTDIR=

# ------------------------------------------------------------------
# 4.  tidy up & show result
# ------------------------------------------------------------------
cd ../..
ls .
rm -rf "$WORK_DIR"
rm -rf "openssl-$OPENSSL_VER"
rm -rf "curl-$CURL_VER"

# unset temp variables
unset CC
unset AR
unset AS
unset NM
unset STRIP
unset RANLIB
unset PKG_CONFIG_PATH


echo "✅ curl + OpenSSL ready:"
tree "$PREFIX"