#!/usr/bin/env bash
# build-bzip2-linux.sh   (native Linux x86-64)
set -euo pipefail

LIB_NAME="bzip2"
BZIP2_VER="1.0.8"
EXT_DIR=".extensions/${LIB_NAME}"
INSTALL_DIR="$(pwd)/bin/linux-x86_64/${LIB_NAME}"
TAR_FILE="/tmp/bzip2-${BZIP2_VER}.tar.gz"
SRC_URL="https://sourceware.org/pub/bzip2/bzip2-${BZIP2_VER}.tar.gz"

# ------------------------------------------------------------------
echo "📦 1. Install build dependencies"
sudo apt-get update -qq
sudo apt-get install -y build-essential curl tar tree

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
echo "🔨 3. Build all native Linux components with -fPIC"
# Clean previous build artifacts
make clean &>/dev/null || true

# STAGE 1: Build the executables and the static library. Critically, we inject
# CFLAGS="-fPIC" so that all object files (.o) are compatible with shared libraries.
make CFLAGS="-O2 -fPIC" -j"$(nproc)"

# STAGE 2: Build the shared library (.so) using the PIC-compatible object
# files that were just created in STAGE 1.
make -f Makefile-libbz2_so

# ------------------------------------------------------------------
echo "📁 4. Install ALL built and documentation artifacts into ${INSTALL_DIR}"
# Clean the destination for a fresh install
rm -rf "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/"{bin,lib,include,share/doc/bzip2,share/man/man1}

# Install executables
cp -f bzip2 bzip2recover bzdiff bzgrep bzmore "${INSTALL_DIR}/bin/"
# On Linux, bunzip2 and bzcat are symlinks to bzip2
(cd "${INSTALL_DIR}/bin" && ln -sf bzip2 bunzip2)
(cd "${INSTALL_DIR}/bin" && ln -sf bzip2 bzcat)

# Install libraries (static and shared)
cp -f libbz2.a "${INSTALL_DIR}/lib/"
cp -f libbz2.so.* "${INSTALL_DIR}/lib/"
# Create standard soname symlinks
(cd "${INSTALL_DIR}/lib" && ln -sf libbz2.so.1.0.8 libbz2.so.1.0)
(cd "${INSTALL_DIR}/lib" && ln -sf libbz2.so.1.0 libbz2.so)

# Install header files
cp -f bzlib.h "${INSTALL_DIR}/include/"

# Install documentation
cp -f CHANGES LICENSE README README.* manual.* "${INSTALL_DIR}/share/doc/bzip2/" 2>/dev/null || true

# Install man pages
cp -f *.1 "${INSTALL_DIR}/share/man/man1/" 2>/dev/null || true

# ------------------------------------------------------------------
echo "🧾 5. Resulting file structure"
tree "${INSTALL_DIR}"

# ------------------------------------------------------------------
echo "✅ Done."