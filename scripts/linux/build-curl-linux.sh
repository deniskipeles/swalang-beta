#!/usr/bin/env bash
# build-curl-linux.sh   (native Linux x86-64)
set -euo pipefail

echo "⚙️  1. Install deps"
sudo apt-get update -qq
sudo apt-get install -y build-essential git autotools-dev libtool \
                        libssl-dev zlib1g-dev libzstd-dev libbrotli-dev \
                        libnghttp2-dev

echo "📥 2. Clone curl"
LIBRARY_NAME="curl"
REPO_DIR=".extensions/$LIBRARY_NAME"
[[ -d "$REPO_DIR" ]] || git clone --depth=1 https://github.com/curl/curl.git "$REPO_DIR"
cd "$REPO_DIR"

echo "🔨 3. Autotools bootstrap & configure"
autoreconf -fi
INSTALL_DIR="$(pwd)/../../bin/linux-x86_64/$LIBRARY_NAME"
./configure --prefix="$INSTALL_DIR" \
            --enable-shared --disable-static \
            --with-ssl --with-zstd --with-brotli \
            --with-nghttp2 --disable-ldap
make -j"$(nproc)"
make install

tree "$INSTALL_DIR"