#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

mkdir -p "$EXT_DIR"
cd "$EXT_DIR"

# ---------------------------------------------------------------------------
# zig rc shim — lives in a mktemp file for the duration of this script only.
# CMake calls the RC compiler as: <RC> <flags> -o <out> <in>
# zig rc expects:                  zig rc <in> -o <out>
# The shim reorders arguments. Cleaned up automatically on EXIT.
# ---------------------------------------------------------------------------
ZIG_RC_SHIM="$(mktemp /tmp/zig-rc-XXXXXX.sh)"
cat > "$ZIG_RC_SHIM" << 'EOF'
#!/bin/bash
set -euo pipefail

out=""
in=""
args=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        /fo|/Fo|-fo|-Fo)
            out="${2:?missing output}"
            shift 2
            ;;
        *)
            in="$1"
            shift
            ;;
    esac
done

# Convert .res → .obj
obj="${out%.res}.obj"

exec zig rc /fo"$obj" -- "$in"
EOF
chmod +x "$ZIG_RC_SHIM"
trap 'rm -f "$ZIG_RC_SHIM"' EXIT

echo "================================================="
echo "🎨 Downloading UI Library Sources..."
echo "================================================="

# 1. Nuklear
mkdir -p nuklear && cd nuklear
if [ ! -f "nuklear.h" ]; then
    wget -q "https://raw.githubusercontent.com/Immediate-Mode-UI/Nuklear/master/nuklear.h"
    cat << 'C_EOF' > nuklear_impl.c
#define NK_IMPLEMENTATION
#include "nuklear.h"
C_EOF
fi
cd ..

# 2. LVGL
if [ ! -d "lvgl" ]; then
    git clone --depth 1 https://github.com/lvgl/lvgl.git lvgl
    cd lvgl
    cp lv_conf_template.h lv_conf.h
    sed -i 's/#if 0/#if 1/g' lv_conf.h
    cd ..
fi

rm -rf lvgl/examples lvgl/demos
find lvgl -name "CMakeLists.txt" \
    -exec sed -i 's/add_library(thorvg SHARED/add_library(thorvg STATIC/g' {} \;

# 3. SDL2
if [ ! -d "sdl2" ]; then
    SDL_VER="2.30.2"
    wget -q "https://github.com/libsdl-org/SDL/releases/download/release-${SDL_VER}/SDL2-${SDL_VER}.tar.gz"
    tar -xf "SDL2-${SDL_VER}.tar.gz"
    mv "SDL2-${SDL_VER}" sdl2
    rm "SDL2-${SDL_VER}.tar.gz"

    find sdl2 -name "version.rc" -delete
    find sdl2 -name "CMakeLists.txt" -exec sed -i '/version.rc/d' {} \;
fi

echo "✔️ UI Sources downloaded and prepared."

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
    local extra_c_flags=""

    if [[ "$target" == *"windows"* ]]; then
        cmake_flags+=("-DCMAKE_SYSTEM_NAME=Windows")
        extra_c_flags="-lws2_32 -lgdi32 -luser32 -limm32 -lole32 -loleaut32 -lshell32 -lversion -luuid -lwinmm"
    else
        extra_c_flags="-pthread -lm"
    fi

    case $lib_name in
        nuklear)
            local out_file="libnuklear.so"
            if [[ "$target" == *"windows"* ]]; then out_file="nuklear.dll"; fi
            zig cc -target "$target" -shared -fPIC -o "$out_dir/$out_file" \
                nuklear_impl.c -O3 $extra_c_flags
            ;;

        lvgl)
            rm -rf build-$target
            local lvgl_flags=(
                "${cmake_flags[@]}"
                "-DBUILD_SHARED_LIBS=ON"
            )
            if [[ "$target" == *"windows"* ]]; then
                lvgl_flags+=(
                    "-DLV_USE_THORVG=0"
                    "-DCONFIG_LV_USE_THORVG_INTERNAL=n"
                    "-DCONFIG_LV_USE_VECTOR_GRAPHIC=n"
                )
            fi
            cmake -B build-$target "${lvgl_flags[@]}"
            cmake --build build-$target --target lvgl --parallel "$(nproc)"
            find build-$target \( -name "liblvgl.so*" -o -name "liblvgl.dll" -o -name "liblvgl.dylib*" \) \
                -exec cp {} "$out_dir/" \;
            ;;

        sdl2)
            rm -rf build-$target
            local sdl_flags=(
                "${cmake_flags[@]}"
                "-DSDL_STATIC=OFF"
                "-DSDL_SHARED=ON"
            )
            if [[ "$target" == *"windows"* ]]; then
                sdl_flags+=(
                    "-DSDL_LIBC=ON"
                    "-DSDL_GCC_ATOMICS=OFF"
                    # ZIG_RC_SHIM is a mktemp file created at script start,
                    # auto-deleted on EXIT — no external wrapper file needed
                    "-DCMAKE_RC_COMPILER=$ZIG_RC_SHIM"
                )
            fi
            cmake -B build-$target "${sdl_flags[@]}"
            cmake --build build-$target --target SDL2 --parallel "$(nproc)"
            find build-$target \( -name "libSDL2*.so*" -o -name "*SDL2*.dll" -o -name "libSDL2*.dylib*" \) \
                -exec cp {} "$out_dir/" \;
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
    echo "   └─ artifacts: $(ls "$out_dir" | tr '\n' ' ')"
}

LIBS=("nuklear" "lvgl" "sdl2")

rm -rf "$BIN_DIR"/*/nuklear "$BIN_DIR"/*/lvgl "$BIN_DIR"/*/sdl2

for target in "${TARGETS[@]}"; do
    for lib in "${LIBS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo ""
echo "✅ All UI shared libraries built successfully → $BIN_DIR"