#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

mkdir -p "$EXT_DIR"
cd "$EXT_DIR"

echo "================================================="
echo "🧮 Downloading Math Library Sources..."
echo "================================================="

# 1. BLIS
if [ ! -d "blis" ]; then
    BLIS_VER="1.0"
    wget -q "https://github.com/flame/blis/archive/refs/tags/${BLIS_VER}.tar.gz" -O blis.tar.gz
    tar -xf blis.tar.gz
    mv "blis-${BLIS_VER}" blis
    rm blis.tar.gz
fi

echo "✔️ Math library sources downloaded and prepared."

build_lib() {
    local lib_name=$1
    local target=$2

    local folder_target="x86_64-linux"
    if [[ "$target" == *"windows"* ]]; then folder_target="x86_64-windows-gnu"; fi

    local out_dir="$BIN_DIR/$folder_target/$lib_name"
    mkdir -p "$out_dir"

    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.dylib 1>/dev/null 2>&1; then
        echo "✔️ $lib_name for $target already exists. Skipping."
        return
    fi

    echo "🔨 Building $lib_name for $target..."
    pushd "$EXT_DIR/$lib_name" > /dev/null

    case $lib_name in
        blis)
            # BLIS uses a custom configure script and is built in-source.
            # We must clean thoroughly between different targets.
            make distclean || true
            rm -rf "install-$target"

            local config_args=("--enable-threading=openmp" "--enable-shared" "--disable-static")

            # Optimization: BLIS generic x86_64
            CC="zig cc -target $target" \
            CXX="zig c++ -target $target" \
            AR="zig ar" \
            RANLIB="zig ranlib" \
            ./configure "${config_args[@]}" --prefix="$PWD/install-$target" generic

            make -j"$(nproc || echo 1)"
            make install

            find "install-$target/lib" \( -name "libblis.so*" -o -name "libblis.dll" -o -name "blis.dll" \) -exec cp {} "$out_dir/" \;
            ;;
    esac

    if [[ "$target" == *"linux"* ]]; then
        pushd "$out_dir" > /dev/null
        for f in *.so.*; do
            [ -e "$f" ] || continue
            base="${f%%.so*}.so"
            [ -L "$base" ] || ln -sf "$f" "$base"
        done
        popd > /dev/null
    fi

    popd > /dev/null
    echo "✅ $lib_name for $target done."
}

LIBS=("blis")

for target in "${TARGETS[@]}"; do
    for lib in "${LIBS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo "🎉 All math shared libraries built successfully!"
