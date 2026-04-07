#!/usr/bin/env bash
# build-sqlite-android-arm64.sh -- builds SQLite3 for Android arm64-v8a
set -euo pipefail

# ------------------------------------------------------------------
# 0.  Android NDK Configuration
# ------------------------------------------------------------------
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
export ANDROID_NDK="$NDK"
HOST_TAG="linux-x86_64"
TOOLCHAIN="$NDK/toolchains/llvm/prebuilt/$HOST_TAG"
export PATH="$TOOLCHAIN/bin:$PATH"

ANDROID_API=24
TARGET_ARCH="arm64"
TARGET_TRIPLE="aarch64-linux-android"

# ------------------------------------------------------------------
# 1.  Directories & Configuration
# ------------------------------------------------------------------
LIB_NAME="sqlite3"
WORK_DIR="$(pwd)/.sqlite-build-tmp"
INSTALL_DIR="$(pwd)/bin/android/arm64-v8a/${LIB_NAME}"

# Check https://www.sqlite.org/download.html for the latest version.
SQLITE_YEAR="2024"
SQLITE_VERSION_CODE="3460000" # Represents version 3.46.0

# Recommended compile-time flags for modern features
SQLITE_COMPILE_FLAGS=(
  "-fPIC"
  "-O2"
  "-DSQLITE_ENABLE_FTS5"
  "-DSQLITE_ENABLE_JSON1"
  "-DSQLITE_ENABLE_RTREE=1"
  "-DSQLITE_DEFAULT_FOREIGN_KEYS=1"
  "-DSQLITE_OMIT_DEPRECATED"
)

mkdir -p "$WORK_DIR" "$INSTALL_DIR"
cd "$WORK_DIR"

# ------------------------------------------------------------------
# 2.  Download & Unpack SQLite Amalgamation Source
# ------------------------------------------------------------------
SQLITE_ZIP_FILE="sqlite-amalgamation-${SQLITE_VERSION_CODE}.zip"
SQLITE_SRC_DIR="sqlite-amalgamation-${SQLITE_VERSION_CODE}"
SQLITE_URL="https://www.sqlite.org/${SQLITE_YEAR}/${SQLITE_ZIP_FILE}"

if [[ ! -f "$SQLITE_ZIP_FILE" ]]; then
  echo "Downloading SQLite ${SQLITE_VERSION_CODE} …"
  curl -fSL "${SQLITE_URL}" -o "${SQLITE_ZIP_FILE}"
fi

if [[ ! -d "$SQLITE_SRC_DIR" ]]; then
    echo "Extracting SQLite source..."
    unzip -q "$SQLITE_ZIP_FILE"
fi

SRC_C_FILE="${SQLITE_SRC_DIR}/sqlite3.c"

# ------------------------------------------------------------------
# 3.  Build Shared Library for Android arm64-v8a
# ------------------------------------------------------------------
echo "Configuring build for Android API ${ANDROID_API} (${TARGET_ARCH})..."

# Set environment variables to point to the NDK's Clang toolchain
export CC="${TOOLCHAIN}/bin/${TARGET_TRIPLE}${ANDROID_API}-clang"
export AR="${TOOLCHAIN}/bin/llvm-ar"
export AS="${TOOLCHAIN}/bin/llvm-as"
export NM="${TOOLCHAIN}/bin/llvm-nm"
export STRIP="${TOOLCHAIN}/bin/llvm-strip"
export RANLIB="${TOOLCHAIN}/bin/llvm-ranlib"

echo "Compiling ${SRC_C_FILE}..."

# FIX: Removed the Linux-specific -lpthread and -ldl flags
$CC -shared "${SQLITE_COMPILE_FLAGS[@]}" \
    -o "${INSTALL_DIR}/lib/libsqlite3.so" \
    "${SRC_C_FILE}"

# ------------------------------------------------------------------
# 4.  Install Headers
# ------------------------------------------------------------------
echo "Installing header files..."
mkdir -p "${INSTALL_DIR}/include"
cp -f "${SQLITE_SRC_DIR}/sqlite3.h"    "${INSTALL_DIR}/include/"
cp -f "${SQLITE_SRC_DIR}/sqlite3ext.h" "${INSTALL_DIR}/include/"

# ------------------------------------------------------------------
# 5.  Tidy Up & Show Result
# ------------------------------------------------------------------
cd ..
rm -rf "$WORK_DIR"

# Unset temporary environment variables
unset CC AR AS NM STRIP RANLIB

echo "✅ SQLite build for Android arm64-v8a complete:"
tree "$INSTALL_DIR"