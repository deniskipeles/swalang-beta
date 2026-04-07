#!/usr/bin/env bash
# build-jansson-android-arm64.sh
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
# 1.  Jansson submodule / clone
# ------------------------------------------------------------------
LIBRARY_NAME="jansson"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 https://github.com/akheron/jansson.git "$REPO_DIR"
cd "$REPO_DIR"

# ------------------------------------------------------------------
# 2.  Clean any stale objects (from previous native builds)
# ------------------------------------------------------------------
make distclean 2>/dev/null || true   # ignore if nothing to clean
git clean -xfd                      # remove every untracked file

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
  --with-pic

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
unset TOOLCHAIN
unset HOST_TAG

e