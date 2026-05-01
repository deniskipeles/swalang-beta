#!/bin/bash
set -euo pipefail

# Use ABSOLUTE paths to prevent pushd/popd relative path errors
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"

# Targeting glibc 2.17 for Linux to ensure maximum portability with dlopen()
TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

mkdir -p "$EXT_DIR"
cd "$EXT_DIR"

echo "================================================="
echo "🎨 Downloading UI Library Sources..."
echo "================================================="

# 1. Nuklear (Single-header immediate mode GUI)
mkdir -p nuklear && cd nuklear
if [ ! -f "nuklear.h" ]; then
    wget -q "https://raw.githubusercontent.com/Immediate-Mode-UI/Nuklear/master/nuklear.h"
    # Create the C implementation wrapper required to compile a shared library from a single-header lib
    cat << 'C_EOF' > nuklear_impl.c
#define NK_IMPLEMENTATION
#include "nuklear.h"
C_EOF
fi
cd ..

# 2. LVGL (Light and Versatile Graphics Library)
if [ ! -d "lvgl" ]; then
    git clone --depth 1 https://github.com/lvgl/lvgl.git lvgl
    cd lvgl
    # LVGL requires a config file. We copy the template and enable it.
    cp lv_conf_template.h lv_conf.h
    sed -i 's/#if 0/#if 1/g' lv_conf.h
    cd ..
fi

# 3. SDL2 (Simple DirectMedia Layer)
if [ ! -d "sdl2" ]; then
    SDL_VER="2.30.2"
    wget -q "https://github.com/libsdl-org/SDL/releases/download/release-${SDL_VER}/SDL2-${SDL_VER}.tar.gz"
    tar -xf "SDL2-${SDL_VER}.tar.gz"
    mv "SDL2-${SDL_VER}" sdl2
    rm "SDL2-${SDL_VER}.tar.gz"
fi

echo "✔️ UI Sources downloaded and prepared."

build_lib() {
    local lib_name=$1
    local target=$2

    local folder_target="x86_64-linux"
    if [[ "$target" == *"windows"* ]]; then folder_target="x86_64-windows-gnu"; fi

    local out_dir="$BIN_DIR/$folder_target/$lib_name"
    mkdir -p "$out_dir"

    # Skip if already built
    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.dylib 1> /dev/null 2>&1; then
        echo "✔️ $lib_name for $target already exists. Skipping."
        return
    fi

    echo "🔨 Building $lib_name for $target..."
    pushd "$EXT_DIR/$lib_name" > /dev/null

    local cmake_flags=(
        "-DCMAKE_C_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_CXX_COMPILER=zig;c++;-target;$target"
        "-DCMAKE_ASM_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_BUILD_TYPE=Release"
    )
    local extra_c_flags=""
    
    if [[ "$target" == *"windows"* ]]; then
        # FIX for Windows: Explicitly export all symbols for DLLs so they behave like Linux .so files!
        cmake_flags+=("-DCMAKE_SYSTEM_NAME=Windows")
        cmake_flags+=("-DCMAKE_SHARED_LINKER_FLAGS=-Xlinker /EXPORTALL") # LLD-specific export all flag
        extra_c_flags="-lws2_32 -lgdi32 -luser32 -limm32 -lole32 -loleaut32 -lshell32 -lversion -luuid -lwinmm"
    else
        extra_c_flags="-pthread -lm"
    fi

    case $lib_name in
        nuklear)
            local out_file="libnuklear.so"
            if [[ "$target" == *"windows"* ]]; then out_file="nuklear.dll"; fi
            # Compile directly using zig cc
            zig cc -target "$target" -shared -fPIC -o "$out_dir/$out_file" nuklear_impl.c -O3 $extra_c_flags
            ;;
        lvgl)
            rm -rf build-$target
            # LVGL builds static by default, so we force shared.
            # We also disable building the examples to save massive compile time and avoid linker bloat!
            local lvgl_flags=("${cmake_flags[@]}" "-DBUILD_SHARED_LIBS=ON" "-DLV_CONF_BUILD_DISABLE_EXAMPLES=ON" "-DLV_CONF_BUILD_DISABLE_DEMOS=ON")
            cmake -B build-$target "${lvgl_flags[@]}"
            cmake --build build-$target --target lvgl --parallel "$(nproc)"
            find build-$target -name "liblvgl.so*" -o -name "*lvgl*.dll" -o -name "liblvgl.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
        sdl2)
            rm -rf build-$target
            # SDL2 specific flags
            local sdl_flags=("${cmake_flags[@]}" "-DSDL_STATIC=OFF" "-DSDL_SHARED=ON")
            if [[ "$target" == *"windows"* ]]; then
                sdl_flags+=("-DSDL_LIBC=ON" "-DSDL_GCC_ATOMICS=OFF")
            fi
            cmake -B build-$target "${sdl_flags[@]}"
            cmake --build build-$target --target SDL2 --parallel "$(nproc)"
            find build-$target -name "libSDL2*.so*" -o -name "*SDL2*.dll" -o -name "libSDL2*.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
    esac
    
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
}

LIBS=("nuklear" "lvgl" "sdl2")

# Force rebuild just in case
rm -rf "$BIN_DIR"/*/nuklear "$BIN_DIR"/*/lvgl "$BIN_DIR"/*/sdl2

for target in "${TARGETS[@]}"; do
    for lib in "${LIBS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo "✅ All UI shared libraries built successfully and placed in $BIN_DIR"