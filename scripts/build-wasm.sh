#!/bin/bash

set -e

# Paths
GO_FILE="./cmd/wasm/main.go"
# OUT_DIR="~/svelte/swalang/static"
# svelte/swalang/static/swalang_wasm
OUT_DIR="$HOME/svelte/swalang/static/swalang_wasm"
WASM_NAME="wasm"
TINYGO_OPTS="-opt=z"
TARGET="wasm"

# Ensure output directory exists
mkdir -p "$OUT_DIR"

echo "🛠️ Building WASM with TinyGo..."
tinygo build -o "$OUT_DIR/${WASM_NAME}.wasm" -target=$TARGET $TINYGO_OPTS "$GO_FILE"

echo "✅ Build complete: ${WASM_NAME}.wasm"

# Check for wasm-opt
if ! command -v wasm-opt &> /dev/null; then
    echo "⚠️  wasm-opt not found. Skipping optimization."
else
    echo "🚀 Optimizing WASM with wasm-opt..."
    wasm-opt -Oz --strip-debug --strip-dwarf --strip-producers \
      -o "$OUT_DIR/${WASM_NAME}_opt.wasm" "$OUT_DIR/${WASM_NAME}.wasm"
    echo "✅ Optimized: ${WASM_NAME}_opt.wasm"
fi

# Gzip compression
echo "📦 Compressing with gzip..."
gzip -kf "$OUT_DIR/${WASM_NAME}_opt.wasm"
echo "✅ Compressed: ${WASM_NAME}_opt.wasm.gz"

# Brotli compression
if command -v brotli &> /dev/null; then
    echo "📦 Compressing with brotli..."
    brotli -f -Z "$OUT_DIR/${WASM_NAME}_opt.wasm"
    echo "✅ Compressed: ${WASM_NAME}_opt.wasm.br"
else
    echo "⚠️  brotli not found. Skipping Brotli compression."
fi

echo "🎉 Done. Output files:"
ls -lh "$OUT_DIR/${WASM_NAME}"*.wasm*
