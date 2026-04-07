#!/bin/bash
set -euo pipefail

ZIG_VERSION="0.13.0"
ZIG_ARCH="x86_64-linux"
ZIG_TARBALL="zig-linux-x86_64-${ZIG_VERSION}.tar.xz"
ZIG_URL="https://ziglang.org/download/${ZIG_VERSION}/${ZIG_TARBALL}"

echo "📥 Downloading Zig ${ZIG_VERSION}..."
curl -L "$ZIG_URL" -o "$ZIG_TARBALL"

echo "📦 Extracting Zig..."
tar -xf "$ZIG_TARBALL"
rm "$ZIG_TARBALL"

# Move to a directory in PATH if possible, or just link it.
# For this environment, we can put it in /usr/local/bin if we have sudo,
# or just add it to the PATH.
ZIG_DIR="zig-linux-x86_64-${ZIG_VERSION}"
sudo mv "$ZIG_DIR" /usr/local/zig
sudo ln -sf /usr/local/zig/zig /usr/local/bin/zig

echo "✅ Zig $(zig version) installed successfully!"
