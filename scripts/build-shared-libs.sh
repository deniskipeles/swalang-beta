#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

TARGETS=("x86_64-linux" "x86_64-windows-gnu")
EXT_DIR=".extensions"
BIN_DIR="bin"

mkdir -p "$EXT_DIR"

declare -A REPOS=(
    [mongoose]="https://github.com/cesanta/mongoose.git"
    [mbedtls]="https://github.com/Mbed-TLS/mbedtls.git"
    [pcre2]="https://github.com/PhilipHazel/pcre2.git"
    [yyjson]="https://github.com/ibireme/yyjson.git"
    [libuv]="https://github.com/libuv/libuv.git"
)

# Clone or update repositories
for lib in "${!REPOS[@]}"; do
    if [ ! -d "$EXT_DIR/$lib" ]; then
        echo "📥 Cloning $lib..."
        git clone --depth 1 "${REPOS[$lib]}" "$EXT_DIR/$lib"
    else
        echo "✔️ $lib already cloned."
    fi
done

build_lib() {
    local lib_name=$1
    local target=$2

    echo "🔨 Building $lib_name for $target..."
    local out_dir="$BIN_DIR/$target/$lib_name"
    mkdir -p "$out_dir"

    pushd "$EXT_DIR/$lib_name" > /dev/null

    local extra_cmake_flags=""
    if [[ "$target" == *"windows"* ]]; then
        extra_cmake_flags="-DCMAKE_SYSTEM_NAME=Windows"
    fi

    case $lib_name in
        mongoose)
            local out_file="libmongoose.so"
            if [[ "$target" == *"windows"* ]]; then out_file="mongoose.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" mongoose.c -D_FILE_OFFSET_BITS=64
            ;;
        yyjson)
            local out_file="libyyjson.so"
            if [[ "$target" == *"windows"* ]]; then out_file="yyjson.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" yyjson.c -I. -O3
            ;;
        pcre2)
            mkdir -p build-zig-"$target"
            pushd build-zig-"$target" > /dev/null
            cmake .. -DCMAKE_C_COMPILER="zig;cc;-target;$target" \
                     -DCMAKE_CXX_COMPILER="zig;c++;-target;$target" \
                     -DCMAKE_BUILD_TYPE=Release \
                     -DBUILD_SHARED_LIBS=ON \
                     -DPCRE2_BUILD_PCRE2_8=ON \
                     -DPCRE2_BUILD_PCRE2_16=OFF \
                     -DPCRE2_BUILD_PCRE2_32=OFF \
                     -DPCRE2_BUILD_TESTS=OFF \
                     -DPCRE2_BUILD_PCRE2GREP=OFF \
                     $extra_cmake_flags
            cmake --build . --parallel "$(nproc)"
            cp *pcre2-8* "$out_dir/"
            popd > /dev/null
            ;;
        mbedtls)
            mkdir -p build-zig-"$target"
            pushd build-zig-"$target" > /dev/null
            cmake .. -DCMAKE_C_COMPILER="zig;cc;-target;$target" \
                     -DCMAKE_CXX_COMPILER="zig;c++;-target;$target" \
                     -DCMAKE_BUILD_TYPE=Release \
                     -DUSE_SHARED_MBEDTLS_LIBRARY=ON \
                     -DENABLE_TESTING=OFF \
                     -DENABLE_PROGRAMS=OFF \
                     $extra_cmake_flags
            cmake --build . --parallel "$(nproc)"
            cp library/libmbed* "$out_dir/" 2>/dev/null || cp library/mbed* "$out_dir/"
            popd > /dev/null
            ;;
        libuv)
            mkdir -p build-zig-"$target"
            pushd build-zig-"$target" > /dev/null
            cmake .. -DCMAKE_C_COMPILER="zig;cc;-target;$target" \
                     -DCMAKE_CXX_COMPILER="zig;c++;-target;$target" \
                     -DCMAKE_BUILD_TYPE=Release \
                     -DBUILD_TESTING=OFF \
                     $extra_cmake_flags
            cmake --build . --parallel "$(nproc)"
            cp libuv* "$out_dir/" 2>/dev/null || cp uv* "$out_dir/"
            popd > /dev/null
            ;;
    esac
    popd > /dev/null
}

for target in "${TARGETS[@]}"; do
    for lib in "${!REPOS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo "✅ Shared libraries built successfully!"
