#!/bin/bash
set -euo pipefail

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Target platforms
TARGETS=("x86_64-linux" "x86_64-windows-gnu")

# Libraries to build
EXT_DIR=".extensions"

# Output base directory
BIN_DIR="bin"

# Function to build a library
build_lib() {
    local lib_name=$1
    local target=$2

    echo "🔨 Building $lib_name for $target..."

    local out_dir="$BIN_DIR/$target/$lib_name"
    mkdir -p "$out_dir"

    local src_dir="$EXT_DIR/$lib_name"
    if [ ! -d "$src_dir" ]; then
        echo "❌ Source directory $src_dir not found. Skipping..."
        return
    fi

    pushd "$src_dir" > /dev/null

    # Extra CMake flags for Windows
    local extra_cmake_flags=""
    if [[ "$target" == *"windows"* ]]; then
        extra_cmake_flags="-DCMAKE_SYSTEM_NAME=Windows"
    fi

    case $lib_name in
        mongoose)
            # Mongoose is easy to build directly
            local out_file="libmongoose.so"
            if [[ "$target" == *"windows"* ]]; then out_file="mongoose.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" mongoose.c -D_FILE_OFFSET_BITS=64
            ;;
        yyjson)
            # yyjson also easy to build directly
            local out_file="libyyjson.so"
            if [[ "$target" == *"windows"* ]]; then out_file="yyjson.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" yyjson.c -I. -O3
            ;;
        pcre2)
            # PCRE2 needs proper configuration, use CMake
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
            # Copy output to out_dir
            cp *pcre2-8* "$out_dir/"
            popd > /dev/null
            ;;
        mbedtls)
            # mbedtls also better with CMake
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
            # Copy output to out_dir
            cp library/libmbed* "$out_dir/" 2>/dev/null || cp library/mbed* "$out_dir/"
            popd > /dev/null
            ;;
        libuv)
            # libuv also better with CMake
            mkdir -p build-zig-"$target"
            pushd build-zig-"$target" > /dev/null
            cmake .. -DCMAKE_C_COMPILER="zig;cc;-target;$target" \
                     -DCMAKE_CXX_COMPILER="zig;c++;-target;$target" \
                     -DCMAKE_BUILD_TYPE=Release \
                     -DBUILD_TESTING=OFF \
                     $extra_cmake_flags
            cmake --build . --parallel "$(nproc)"
            # Copy output to out_dir
            cp libuv* "$out_dir/" 2>/dev/null || cp uv* "$out_dir/"
            popd > /dev/null
            ;;
    esac

    popd > /dev/null
}

for target in "${TARGETS[@]}"; do
    for lib in mongoose yyjson pcre2 mbedtls libuv; do
        build_lib "$lib" "$target"
    done
done

echo "✅ All libraries built successfully!"
