#!/usr/bin/env bash
# build-libyaml-android-arm64.sh
set -euo pipefail

# ------------------------------------------------------------------
# 0.  One-shot CMake 3.31 (re-use the same folder as cJSON build)
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
# 1.  Android NDK
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
# 2.  libyaml submodule
# ------------------------------------------------------------------
LIBRARY_NAME="libyaml"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 https://github.com/yaml/libyaml.git "$REPO_DIR"
cd "$REPO_DIR"

# ------------------------------------------------------------------
# 3.  Android arm64-v8a build
# ------------------------------------------------------------------
INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"

cmake -B build -G Ninja \
  -DCMAKE_TOOLCHAIN_FILE="$NDK_DIR/build/cmake/android.toolchain.cmake" \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-24 \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DYAML_STATIC_LIB_NAME=yaml \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5   # allow old syntax

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"