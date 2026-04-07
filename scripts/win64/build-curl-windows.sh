#!/usr/bin/env bash
# build-curl-windows.sh  –  cross-compile curl for Windows x64 (MinGW-w64)
set -euo pipefail

# ----------------------------------------------------------
echo "📦 1. Install MinGW-w64 + build deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git mingw-w64 \
                        libz-mingw-w64-dev libssl-dev \
                        autoconf libtool pkg-config

# ----------------------------------------------------------
echo "🧬 2. Clone official curl source"
LIBRARY_NAME="curl"
REPO_DIR=".extensions/$LIBRARY_NAME"
if [[ ! -d "$REPO_DIR" ]]; then
    git clone --depth=1 https://github.com/curl/curl.git "$REPO_DIR"
fi
cd "$REPO_DIR"

# ----------------------------------------------------------
echo "🔨 3. Configure for Windows x64 cross-compile"
INSTALL_DIR="$(pwd)/../../bin/windows-x86_64/$LIBRARY_NAME"
HOST="x86_64-w64-mingw32"
export PKG_CONFIG_PATH="/usr/${HOST}/lib/pkgconfig"

./buildconf               # generate configure script
./configure \
  --host=x86_64-w64-mingw32 \
  --prefix="${INSTALL_DIR}" \
  --with-zlib \
  --enable-shared \
  --enable-static \
  --without-ssl \
  --disable-ldap \
  --without-libpsl \
  --without-brotli \
  --without-zstd \
  --without-nghttp2 \
  --disable-ipv6   # optional: if you don’t need IPv6
# ./configure \
#     --host="${HOST}" \
#     --prefix="${INSTALL_DIR}" \
#     --with-ssl \
#     --with-zlib \
#     --enable-shared \
#     --enable-static \
#     --disable-ldap \
#     --disable-ldaps

# ----------------------------------------------------------
echo "⚙️  4. Build & install"
make -j"$(nproc)"
make install

# ----------------------------------------------------------
echo "🧾 5. Result"
tree "${INSTALL_DIR}"