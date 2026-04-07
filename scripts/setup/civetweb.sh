#!/bin/bash
# ./scripts/setup_civetweb.sh

# Create the .extensions directory if it doesn't exist
mkdir -p .extensions

# Civetweb C library
LIB_NAME="civetweb"
LIB_URL="https://github.com/civetweb/civetweb.git"
TARGET_DIR=".extensions/$LIB_NAME"

echo "Setting up $LIB_NAME submodule..."

if [ -d "$TARGET_DIR" ]; then
    echo "✔ $LIB_NAME already exists at $TARGET_DIR"
else
    echo "➕ Adding $LIB_NAME..."
    git submodule add --depth=1 "$LIB_URL" "$TARGET_DIR"
fi

echo "✅ Done! Now run:"
echo "   git submodule update --init --recursive"