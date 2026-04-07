#!/usr/bin/env bash
# build-bzip2-windows.sh  –  cross-compile bzip2 for Windows x64 (MinGW-w64)
set -euo pipefail

LIB_NAME="bzip2"
BZIP2_VER="1.0.8"
EXT_DIR=".extensions/${LIB_NAME}"
INSTALL_DIR="$(pwd)/bin/windows-x86_64/${LIB_NAME}"
TAR_FILE="/tmp/bzip2-${BZIP2_VER}.tar.gz"
SRC_URL="https://sourceware.org/pub/bzip2/bzip2-${BZIP2_VER}.tar.gz"

# ------------------------------------------------------------------
echo "📦 1. Install MinGW-w64 + build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential mingw-w64 curl tar tree

# ------------------------------------------------------------------
echo "🧬 2. Download & unpack source if it doesn't exist"
if [[ ! -d "${EXT_DIR}" ]]; then
  echo "-> Source directory '${EXT_DIR}' not found. Downloading..."
  mkdir -p "${EXT_DIR}"
  curl -L "${SRC_URL}" -o "${TAR_FILE}"
  tar -xzf "${TAR_FILE}" -C "${EXT_DIR}" --strip-components=1
  rm "${TAR_FILE}"
else
  echo "-> Source directory '${EXT_DIR}' already exists. Skipping download."
fi
cd "${EXT_DIR}"

# ------------------------------------------------------------------
echo "🔨 3. Build all Windows x64 components"
# Clean previous artifacts to ensure a fresh build
make clean &>/dev/null || true

HOST="x86_64-w64-mingw32"

# A) Compile object files for the library with Position-Independent Code
$HOST-gcc -fPIC -O2 -D_WIN32 -c \
    blocksort.c huffman.c crctable.c randtable.c \
    compress.c decompress.c bzlib.c

# B) Compile object files for the executables
$HOST-gcc -O2 -D_WIN32 -c bzip2.c bzip2recover.c

# C) Create the static library (.a)
$HOST-ar rcs libbz2.a \
    blocksort.o huffman.o crctable.o randtable.o \
    compress.o decompress.o bzlib.o

# D) Create the shared library (.dll) and its import library (.dll.a)
$HOST-gcc -shared -o libbz2.dll -Wl,--out-implib,libbz2.dll.a \
    blocksort.o huffman.o crctable.o randtable.o \
    compress.o decompress.o bzlib.o

# E) Create the final executables by linking them to the new library
$HOST-gcc bzip2.o -o bzip2.exe -L. -lbz2
# bzip2recover is standalone and does not link against libbz2
$HOST-gcc bzip2recover.o -o bzip2recover.exe

# ------------------------------------------------------------------
echo "📁 4. Install ALL built and documentation artifacts into ${INSTALL_DIR}"
# Clean the destination directory for a fresh install
rm -rf "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/"{bin,lib,include,share/doc/bzip2,share/man/man1}

# Install executables and runtime libraries
cp -f bzip2.exe bzip2recover.exe "${INSTALL_DIR}/bin/"
cp -f libbz2.dll                  "${INSTALL_DIR}/bin/"

# Install link-time libraries (static and import)
cp -f libbz2.a      "${INSTALL_DIR}/lib/"
cp -f libbz2.dll.a  "${INSTALL_DIR}/lib/"

# Install header files
cp -f bzlib.h       "${INSTALL_DIR}/include/"

# Install documentation
cp -f CHANGES LICENSE README README.* manual.* "${INSTALL_DIR}/share/doc/bzip2/" 2>/dev/null || true

# Install man pages for completeness
cp -f *.1 "${INSTALL_DIR}/share/man/man1/" 2>/dev/null || true

# ------------------------------------------------------------------
echo "🧾 5. Resulting file structure"
tree "${INSTALL_DIR}"

# ------------------------------------------------------------------
echo "✅ Done."