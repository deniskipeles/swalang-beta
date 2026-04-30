#!/bin/bash
set -euo pipefail

# Use ABSOLUTE paths to prevent pushd/popd relative path errors
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

mkdir -p "$EXT_DIR"
cd "$EXT_DIR"

echo "================================================="
echo "🛡️  Building Isolated MbedTLS FFI Dependency"
echo "================================================="

# 1. Safely clone MbedTLS with all its nested crypto submodules
if [ ! -d "mbedtls" ]; then
    echo "📥 Cloning MbedTLS and submodules..."
    git clone --depth 1 --recurse-submodules https://github.com/Mbed-TLS/mbedtls.git mbedtls
else
    echo "✔️ MbedTLS source already exists. Updating submodules just in case..."
    pushd mbedtls > /dev/null
    git submodule update --init --recursive || true
    popd > /dev/null
fi

# 2. Build for each target
for target in "${TARGETS[@]}"; do
    folder_target="x86_64-linux"
    if [[ "$target" == *"windows"* ]]; then folder_target="x86_64-windows-gnu"; fi

    out_dir="$BIN_DIR/$folder_target/mbedtls"
    mkdir -p "$out_dir"

    # Skip if already built
    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.dylib 1> /dev/null 2>&1; then
        echo "✔️ mbedtls for $target already exists. Skipping."
        continue
    fi

    echo "🔨 Compiling mbedtls for $target..."
    pushd mbedtls > /dev/null

    # Use a Bash Array for safe argument passing (preserves semicolons and spaces!)
    cmake_flags=(
        "-DCMAKE_C_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_CXX_COMPILER=zig;c++;-target;$target"
        "-DCMAKE_ASM_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_BUILD_TYPE=MinSizeRel"
        "-DUSE_SHARED_MBEDTLS_LIBRARY=ON"
        "-DENABLE_TESTING=OFF"
        "-DENABLE_PROGRAMS=OFF"
        "-DMBEDTLS_FATAL_WARNINGS=OFF"
    )

    # Platform-specific tweaks
    if [[ "$target" == *"windows"* ]]; then
        cmake_flags+=("-DCMAKE_SYSTEM_NAME=Windows")
        # Use -Xlinker to forcefully tell LLD to ignore duplicate ___chkstk_ms symbols
        cmake_flags+=("-DCMAKE_SHARED_LINKER_FLAGS=-Xlinker /force:multiple")
        cmake_flags+=("-DCMAKE_C_FLAGS=-mno-stack-arg-probe")
    fi

    rm -rf build-$target
    
    # Execute CMake using the array expansion to perfectly preserve argument boundaries
    cmake -B build-$target "${cmake_flags[@]}"
    cmake --build build-$target --parallel "$(nproc)"

    # Extract all generated DLLs/SOs
    find build-$target -name "libmbed*.so*" -o -name "mbed*.dll" -o -name "libmbed*.dylib*" -o -name "libtfpsacrypto*.dll" | xargs -I {} cp {} "$out_dir/"

    # Symlink cleanup for Linux
    if [[ "$target" == *"linux"* ]]; then
        pushd "$out_dir" > /dev/null
        for f in *.so.*; do
            [ -e "$f" ] || continue
            base="${f%%.so*}.so"
            ln -sf "$f" "$base"
        done
        popd > /dev/null
    fi

    popd > /dev/null
    echo "✅ Success for $target"
done

echo "🎉 Isolated MbedTLS build complete!"