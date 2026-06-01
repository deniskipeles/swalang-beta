#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

mkdir -p "$EXT_DIR"
cd "$EXT_DIR"

echo "================================================="
echo "🎮 Downloading Game Library Sources..."
echo "================================================="

# 1. cglm
if [ ! -d "cglm" ]; then
    CGLM_VER="0.9.6"
    wget -q "https://github.com/recp/cglm/archive/refs/tags/v${CGLM_VER}.tar.gz" -O cglm.tar.gz
    tar -xf cglm.tar.gz
    mv "cglm-${CGLM_VER}" cglm
    rm cglm.tar.gz
fi

# 2. miniaudio
mkdir -p miniaudio && cd miniaudio
if [ ! -f "miniaudio.h" ]; then
    MINIAUDIO_VER="0.11.25"
    wget -q "https://raw.githubusercontent.com/mackron/miniaudio/v${MINIAUDIO_VER}/miniaudio.h"
fi

cat << 'C_EOF' > miniaudio_impl.c
#define MINIAUDIO_IMPLEMENTATION
#include "miniaudio.h"
C_EOF
cd ..

# 3. Box2D
if [ ! -d "box2d" ]; then
    BOX2D_VER="3.1.1"
    wget -q "https://github.com/erincatto/box2d/archive/refs/tags/v${BOX2D_VER}.tar.gz" -O box2d.tar.gz
    tar -xf box2d.tar.gz
    mv "box2d-${BOX2D_VER}" box2d
    rm box2d.tar.gz
fi

# 4. Jolt Physics
if [ ! -d "jolt" ]; then
    JOLT_VER="5.5.0"
    wget -q "https://github.com/jrouwe/JoltPhysics/archive/refs/tags/v${JOLT_VER}.tar.gz" -O jolt.tar.gz
    tar -xf jolt.tar.gz
    mv "JoltPhysics-${JOLT_VER}" jolt
    rm jolt.tar.gz
fi

# 5. Bullet3
if [ ! -d "bullet" ]; then
    BULLET_VER="3.25"
    wget -q "https://github.com/bulletphysics/bullet3/archive/refs/tags/${BULLET_VER}.tar.gz" -O bullet.tar.gz
    tar -xf bullet.tar.gz
    mv "bullet3-${BULLET_VER}" bullet
    rm bullet.tar.gz
fi

echo "✔️ Game library sources downloaded and prepared."

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

    local cmake_flags=(
        "-DCMAKE_C_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_CXX_COMPILER=zig;c++;-target;$target"
        "-DCMAKE_ASM_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_BUILD_TYPE=Release"
    )

    if [[ "$target" == *"windows"* ]]; then
        cmake_flags+=("-DCMAKE_SYSTEM_NAME=Windows")
    fi

    case $lib_name in
        cglm)
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" -DCGLM_STATIC=OFF -DCGLM_SHARED=ON
            cmake --build build-$target --parallel "$(nproc || echo 1)"
            find build-$target \( -name "libcglm.so*" -o -name "cglm.dll" \) -exec cp {} "$out_dir/" \;
            ;;
        miniaudio)
            local out_file="libminiaudio.so"
            local extra_flags="-lpthread -lm -ldl"
            if [[ "$target" == *"windows"* ]]; then
                out_file="miniaudio.dll"
                extra_flags="-lwinmm"
            fi
            zig cc -target "$target" -shared -fPIC -o "$out_dir/$out_file" miniaudio_impl.c -O3 $extra_flags
            ;;
        box2d)
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" -DBOX2D_BUILD_SHARED=ON -DBOX2D_BUILD_SAMPLES=OFF -DBOX2D_BUILD_UNIT_TESTS=OFF
            cmake --build build-$target --parallel "$(nproc || echo 1)"
            find build-$target \( -name "libbox2d.so*" -o -name "box2d.dll" \) -exec cp {} "$out_dir/" \;
            ;;
        jolt)
            rm -rf build-$target
            # Jolt needs some specific flags.
            # TARGET_UNIT_TESTS=OFF, TARGET_SAMPLES=OFF
            # We want it as a shared library.
            cmake -B build-$target -S Build "${cmake_flags[@]}" \
                -DTARGET_UNIT_TESTS=OFF \
                -DTARGET_SAMPLES=OFF \
                -DGENERATE_DEBUG_SYMBOLS=OFF \
                -DCMAKE_POSITION_INDEPENDENT_CODE=ON

            # Jolt's CMakeLists.txt might not support BUILD_SHARED_LIBS=ON directly for everything.
            # We might need to force it or check how Jolt builds shared libs.
            # Jolt usually builds static libs by default.

            cmake --build build-$target --parallel "$(nproc || echo 1)"

            # Jolt by default produces static libs. If we want shared, we might need a wrapper or to patch it.
            # Let's see if we can just link all objects into a shared lib if it didn't make one.
            if ! ls build-$target/*.so build-$target/*.dll 1>/dev/null 2>&1; then
                echo "⚠️ Jolt didn't produce a shared library. Attempting to link static libs into shared..."
                local jolt_libs=$(find build-$target -name "*.a" -o -name "*.lib")
                if [[ "$target" == *"windows"* ]]; then
                    zig c++ -target "$target" -shared -o "$out_dir/jolt.dll" -Wl,--whole-archive $jolt_libs -Wl,--no-whole-archive
                else
                    zig c++ -target "$target" -shared -o "$out_dir/libjolt.so" -Wl,--whole-archive $jolt_libs -Wl,--no-whole-archive
                fi
            else
                find build-$target \( -name "libJolt.so*" -o -name "Jolt.dll" \) -exec cp {} "$out_dir/" \;
            fi
            ;;
        bullet)
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" \
                -DBUILD_SHARED_LIBS=ON \
                -DBUILD_UNIT_TESTS=OFF \
                -DBUILD_BULLET2_DEMOS=OFF \
                -DBUILD_CPU_DEMOS=OFF \
                -DBUILD_OPENGL3_DEMOS=OFF \
                -DBUILD_BULLET_ROBOTICS_GUI_HELPER=OFF \
                -DBUILD_PYBULLET=OFF
            cmake --build build-$target --parallel "$(nproc || echo 1)"
            # Bullet produces many shared libs. Let's grab the main ones.
            # Using -exec instead of grep/xargs for better robustness.
            find build-$target -name "*.so*" -o -name "*.dll" -not -name "*demo*" -exec cp {} "$out_dir/" \;
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

LIBS=("cglm" "miniaudio" "box2d" "jolt" "bullet")

for target in "${TARGETS[@]}"; do
    for lib in "${LIBS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo "🎉 All game shared libraries built successfully!"
