#!/usr/bin/env bash
# build-cjson-linux.sh   (native Linux x86-64)
set -euo pipefail

# ------------------------------------------------------------------
# 0.  One-shot, throw-away CMake 3.31 for this build only
# ------------------------------------------------------------------
TEMP_CMAKE="$HOME/.local/cmake-3.31"
mkdir -p "$TEMP_CMAKE"

if [[ ! -x "$TEMP_CMAKE/bin/cmake" ]]; then
  echo "⚙️  Fetching temporary CMake 3.31 …"
  curl -sL https://github.com/Kitware/CMake/releases/download/v3.31.6/cmake-3.31.6-linux-x86_64.tar.gz \
    | tar -xz -C "$TEMP_CMAKE" --strip-components=1
fi
# Put it first in PATH for this script only
export PATH="$TEMP_CMAKE/bin:$PATH"

# Sanity check
echo "Using $(cmake --version | head -1)"

# ------------------------------------------------------------------
# 1.  Native build dependencies (already satisfied in most cases)
# ------------------------------------------------------------------
echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git

# ------------------------------------------------------------------
# 2.  Clone / update cJSON submodule
# ------------------------------------------------------------------
echo "📥 2. Clone / update cJSON"
LIBRARY_NAME="cjson"
REPO_DIR=".extensions/$LIBRARY_NAME"
git submodule update --init --depth=1 -- "$REPO_DIR"
cd "$REPO_DIR"

# ------------------------------------------------------------------
# 3.  CMake build (shared)
# ------------------------------------------------------------------
echo "🔨 3. CMake build (shared)"
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"

cmake -B build \
  -DCMAKE_BUILD_TYPE=Release \
  -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
  -DBUILD_SHARED_LIBS=ON \
  -DENABLE_CJSON_TEST=OFF

cmake --build build -j"$(nproc)"
cmake --install build

echo "✅ 4. Installed artifacts:"
tree "$INSTALL_DIR"