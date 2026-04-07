#!/usr/bin/env bash
# build-libffi-android-arm64.sh
# Builds libffi (https://github.com/libffi/libffi) for Android arm64-v8a
# Shared library only, no docs, no tests.
set -euo pipefail

# ------------------------------------------------------------------
# 0.  Android NDK (re-use the one you already have)
# ------------------------------------------------------------------
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
HOST_TAG="linux-x86_64"
TOOLCHAIN="$NDK/toolchains/llvm/prebuilt/$HOST_TAG"
export ANDROID_NDK="$NDK"
export PATH="$TOOLCHAIN/bin:$PATH"
API=24
TRIPLE="aarch64-linux-android"

# ------------------------------------------------------------------
# 1.  libffi submodule / clone
# ------------------------------------------------------------------
LIBRARY_NAME="libffi"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 --branch v3.4.6 \
  https://github.com/libffi/libffi.git "$REPO_DIR"
cd "$REPO_DIR"

# ------------------------------------------------------------------
# 2.  Clean any stale objects
# ------------------------------------------------------------------
make distclean 2>/dev/null || true
git clean -xfd

# ------------------------------------------------------------------
# 3.  Autotools cross-build for Android arm64-v8a
# ------------------------------------------------------------------
INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
mkdir -p "$INSTALL_DIR"

export CC="${TOOLCHAIN}/bin/${TRIPLE}${API}-clang"
export CXX="${TOOLCHAIN}/bin/${TRIPLE}${API}-clang++"
export AR="${TOOLCHAIN}/bin/llvm-ar"
export AS="${TOOLCHAIN}/bin/llvm-as"
export NM="${TOOLCHAIN}/bin/llvm-nm"
export STRIP="${TOOLCHAIN}/bin/llvm-strip"
export RANLIB="${TOOLCHAIN}/bin/llvm-ranlib"

[[ -f configure ]] || autoreconf -fi

./configure --host="aarch64-linux-android" \
  --prefix="$INSTALL_DIR" \
  --disable-static \
  --enable-shared \
  --with-pic \
  --disable-docs \
  --disable-multi-os-directory

make -j"$(nproc)"
make install

# ------------------------------------------------------------------
# 4.  show result
# ------------------------------------------------------------------
echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"

# unset temp variables
unset CC
unset CXX
unset AR
unset AS
unset NM
unset STRIP
unset RANLIB

echo "✅ Done."