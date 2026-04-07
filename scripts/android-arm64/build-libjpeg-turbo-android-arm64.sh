#!/usr/bin/env bash
# build-libjpeg-turbo-android-arm64.sh
# Cross-compile libjpeg-turbo for Android arm64-v8a
set -euo pipefail

# ------------------------------------------------------------------
# 0.  One-shot CMake 3.31 (same folder you already use)
# ------------------------------------------------------------------
TEMP_CMAKE="$HOME/.local/cmake-3.31"
mkdir -p "$TEMP_CMAKE"
if [[ ! -x "$TEMP_CMAKE/bin/cmake" ]]; then
  echo "⚙️  Fetching temporary CMake 3.31 …"
  curl -sL https://github.com/Kitware/CMake/releases/download/v3.31.6/cmake-3.31.6-linux-x86_64.tar.gz \
    | tar -xz -C "$TEMP_CMAKE" --strip-components=1
fi
export PATH="$TEMP_CMAKE/bin:$PATH"
echo "Using $(cmake --version | head -1)"

# ------------------------------------------------------------------
# 1.  Android NDK (re-use or auto-download)
# ------------------------------------------------------------------
NDK_DIR="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
if [[ ! -d "$NDK_DIR" ]]; then
  echo "Downloading Android NDK r26d …"
  wget -q --show-progress \
    https://dl.google.com/android/repository/android-ndk-r26d-linux.zip \
    -O "$HOME/android-ndk-r26d-linux.zip"
  unzip -q "$HOME/android-ndk-r26d-linux.zip" -d "$HOME"
  rm "$HOME/android-ndk-r26d-linux.zip"
fi

# ------------------------------------------------------------------
# 2.  libjpeg-turbo submodule / clone
# ------------------------------------------------------------------
LIBRARY_NAME="libjpeg-turbo"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --branch 2.1.5.1 --depth=1 https://github.com/libjpeg-turbo/libjpeg-turbo.git "$REPO_DIR"
cd "$REPO_DIR"

# ------------------------------------------------------------------
# 3.  CMake cross-build for Android arm64-v8a
# ------------------------------------------------------------------
INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
rm -rf build

cmake -B build -G Ninja \
  -DCMAKE_TOOLCHAIN_FILE="$NDK_DIR/build/cmake/android.toolchain.cmake" \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-24 \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DENABLE_STATIC=OFF \
  -DWITH_JPEG8=ON \
  -DWITH_SIMD=ON \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"