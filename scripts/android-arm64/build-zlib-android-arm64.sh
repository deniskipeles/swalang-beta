#!/usr/bin/env bash
# build-zlib-android-arm64.sh
set -euo pipefail

TEMP_CMAKE="$HOME/.local/cmake-3.31"
mkdir -p "$TEMP_CMAKE"
[[ -x "$TEMP_CMAKE/bin/cmake" ]] || \
  curl -sL https://github.com/Kitware/CMake/releases/download/v3.31.6/cmake-3.31.6-linux-x86_64.tar.gz | tar -xz -C "$TEMP_CMAKE" --strip-components=1
export PATH="$TEMP_CMAKE/bin:$PATH"

NDK_DIR="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
if [[ ! -d "$NDK_DIR" ]]; then
  wget -q --show-progress https://dl.google.com/android/repository/android-ndk-r26d-linux.zip -O "$HOME/android-ndk-r26d-linux.zip"
  unzip -q "$HOME/android-ndk-r26d-linux.zip" -d "$HOME" && rm "$HOME/android-ndk-r26d-linux.zip"
fi

LIBRARY_NAME="zlib"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 --branch v1.3.1 https://github.com/madler/zlib.git "$REPO_DIR"
cd "$REPO_DIR"

INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
rm -rf build

cmake -B build \
  -DCMAKE_TOOLCHAIN_FILE="$NDK_DIR/build/cmake/android.toolchain.cmake" \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-24 \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DZLIB_BUILD_EXAMPLES=OFF \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"


# unset temp variables
# unset CC
# unset CXX
# unset AR
# unset AS
# unset NM
# unset STRIP
# unset RANLIB

echo "✅ Done."