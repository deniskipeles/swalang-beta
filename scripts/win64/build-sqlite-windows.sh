#!/usr/bin/env bash
# build-sqlite-windows.sh -- cross-compiles the SQLite3 amalgamation for Windows x64.
set -euo pipefail

# --- Configuration ---
LIB_NAME="sqlite3"
# Check https://www.sqlite.org/download.html for the latest version.
SQLITE_YEAR="2024"
SQLITE_VERSION_CODE="3460000" # Represents version 3.46.0

EXT_DIR=".extensions/sqlite"
INSTALL_DIR="$(pwd)/bin/windows-x86_64/${LIB_NAME}"
SRC_C_FILE="${EXT_DIR}/sqlite3.c"
SRC_H_FILE="${EXT_DIR}/sqlite3.h"

# Recommended compile-time flags to enable modern features
SQLITE_COMPILE_FLAGS=(
  "-fPIC"
  "-O2"
  "-DSQLITE_ENABLE_FTS5"
  "-DSQLITE_ENABLE_JSON1"
  "-DSQLITE_ENABLE_RTREE=1"
  "-DSQLITE_DEFAULT_FOREIGN_KEYS=1"
  "-DSQLITE_OMIT_DEPRECATED"
)

# --- Script ---
echo "📦 1. Install MinGW-w64 + build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential mingw-w64 curl unzip tree

# ------------------------------------------------------------------
echo "🧬 2. Ensure SQLite source amalgamation is present"
# Create the extensions directory if it doesn't exist
mkdir -p "${EXT_DIR}"

if [[ ! -f "${SRC_C_FILE}" ]]; then
    echo "-> sqlite3.c not found. Downloading from sqlite.org..."
    
    SQLITE_ZIP_FILE="sqlite-amalgamation-${SQLITE_VERSION_CODE}.zip"
    SQLITE_URL="https://www.sqlite.org/${SQLITE_YEAR}/${SQLITE_ZIP_FILE}"
    DOWNLOAD_PATH="/tmp/${SQLITE_ZIP_FILE}"
    EXTRACT_TMP_DIR="/tmp/sqlite-amalgamation-tmp"

    echo "   Downloading from ${SQLITE_URL}"
    curl -fSL "${SQLITE_URL}" -o "${DOWNLOAD_PATH}"

    echo "   Extracting..."
    rm -rf "${EXTRACT_TMP_DIR}"
    unzip -q "${DOWNLOAD_PATH}" -d "${EXTRACT_TMP_DIR}"

    # The extracted dir is named like 'sqlite-amalgamation-3460000'
    EXTRACTED_CONTENT_DIR="${EXTRACT_TMP_DIR}/sqlite-amalgamation-${SQLITE_VERSION_CODE}"

    echo "   Copying source files to ${EXT_DIR}/"
    cp -f "${EXTRACTED_CONTENT_DIR}/sqlite3.c" "${EXT_DIR}/"
    cp -f "${EXTRACTED_CONTENT_DIR}/sqlite3.h" "${EXT_DIR}/"
    cp -f "${EXTRACTED_CONTENT_DIR}/sqlite3ext.h" "${EXT_DIR}/"

    echo "   Cleaning up temporary files..."
    rm -f "${DOWNLOAD_PATH}"
    rm -rf "${EXTRACT_TMP_DIR}"
    
    echo "✅ Source download complete."
else
    echo "-> Found existing SQLite source files in ${EXT_DIR}. Skipping download."
fi

# ------------------------------------------------------------------
echo "🔨 3. Cross-compile shared library for Windows x64"
rm -rf "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/bin" "${INSTALL_DIR}/lib" "${INSTALL_DIR}/include"

x86_64-w64-mingw32-gcc -shared "${SQLITE_COMPILE_FLAGS[@]}" \
    -o "${INSTALL_DIR}/bin/libsqlite3.dll" \
    -Wl,--out-implib,"${INSTALL_DIR}/lib/libsqlite3.dll.a" \
    "${SRC_C_FILE}"

# ------------------------------------------------------------------
echo "📁 4. Install header files"
cp -f "${EXT_DIR}/sqlite3.h"    "${INSTALL_DIR}/include/"
cp -f "${EXT_DIR}/sqlite3ext.h" "${INSTALL_DIR}/include/"

# ------------------------------------------------------------------
echo "🧾 5. Resulting file structure"
tree "${INSTALL_DIR}"

# ------------------------------------------------------------------
echo "✅ Done."