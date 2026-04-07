#!/usr/bin/env bash
# build-pcre2-android-arm64.sh
set -euo pipefail

TEMP_CMAKE="$HOME/.local/cmake-3.31"
mkdir -p "$TEMP_CMAKE"
if [[ ! -x "$TEMP_CMAKE/bin/cmake" ]]; then
  echo "⚙️  Fetching temporary CMake 3.31 …"
  curl -sL https://github.com/Kitware/CMake/releases/download/v3.31.6/cmake-3.31.6-linux-x86_64.tar.gz \
    | tar -xz -C "$TEMP_CMAKE" --strip-components=1
fi
export PATH="$TEMP_CMAKE/bin:$PATH"

NDK_DIR="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
if [[ ! -d "$NDK_DIR" ]]; then
  echo "Downloading Android NDK r26d …"
  wget -q --show-progress \
    https://dl.google.com/android/repository/android-ndk-r26d-linux.zip \
    -O "$HOME/android-ndk-r26d-linux.zip"
  unzip -q "$HOME/android-ndk-r26d-linux.zip" -d "$HOME"
  rm "$HOME/android-ndk-r26d-linux.zip"
fi

LIBRARY_NAME="pcre2"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth 1 --branch pcre2-10.44 \
  https://github.com/PCRE2Project/pcre2.git "$REPO_DIR"
cd "$REPO_DIR"

INSTALL_DIR="$(pwd)/../../bin/android/arm64-v8a/$LIBRARY_NAME"
rm -rf build

# ------------------------------------------------------------------
# 3.  Create a tiny CMake file that kills the version-script
# ------------------------------------------------------------------
cat > pcre2-android-fix.cmake <<'EOF'
# Run after project() but before any add_library()
set(CMAKE_C_VISIBILITY_PRESET default)
set(CMAKE_CXX_VISIBILITY_PRESET default)
# Override the variable that holds the script name -> empty
set(PCRE2_VERSION_SCRIPT "")
EOF

cmake -B build \
  -DCMAKE_TOOLCHAIN_FILE="$NDK_DIR/build/cmake/android.toolchain.cmake" \
  -DANDROID_ABI=arm64-v8a \
  -DANDROID_PLATFORM=android-24 \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DPCRE2_BUILD_TESTS=OFF \
  -DPCRE2_BUILD_PCRE2GREP=OFF \
  -DPCRE2_SUPPORT_JIT=ON \
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
  -C pcre2-android-fix.cmake          # <- inject the fix

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ Installed artefacts:"
tree "$INSTALL_DIR"