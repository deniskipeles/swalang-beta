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

echo "📥 Downloading vendor library sources..."

# 1. Mongoose
mkdir -p mongoose && cd mongoose
if [ ! -f "mongoose.c" ]; then
    wget -q "https://raw.githubusercontent.com/cesanta/mongoose/master/mongoose.c"
    wget -q "https://raw.githubusercontent.com/cesanta/mongoose/master/mongoose.h"
fi
cd ..

# 2. YYJSON
mkdir -p yyjson && cd yyjson
if [ ! -f "yyjson.c" ]; then
    wget -q "https://raw.githubusercontent.com/ibireme/yyjson/master/src/yyjson.c"
    wget -q "https://raw.githubusercontent.com/ibireme/yyjson/master/src/yyjson.h"
fi
cd ..

# 3. SQLite3
if [ ! -d "sqlite3" ]; then
    wget -q "https://www.sqlite.org/2024/sqlite-autoconf-3450300.tar.gz"
    tar -xf sqlite-autoconf-3450300.tar.gz
    mv sqlite-autoconf-3450300 sqlite3
    rm sqlite-autoconf-3450300.tar.gz
fi

# 4. Zlib
if [ ! -d "zlib" ]; then
    wget -q "https://github.com/madler/zlib/releases/download/v1.3.1/zlib-1.3.1.tar.gz"
    tar -xf zlib-1.3.1.tar.gz
    mv zlib-1.3.1 zlib
    rm zlib-1.3.1.tar.gz
fi

# 5. Zstd
if [ ! -d "zstd" ]; then
    wget -q "https://github.com/facebook/zstd/releases/download/v1.5.6/zstd-1.5.6.tar.gz"
    tar -xf zstd-1.5.6.tar.gz
    mv zstd-1.5.6 zstd
    rm zstd-1.5.6.tar.gz
fi

# 6. PCRE2
if [ ! -d "pcre2" ]; then
    PCRE2_VER="10.43"
    wget -q "https://github.com/PCRE2Project/pcre2/releases/download/pcre2-${PCRE2_VER}/pcre2-${PCRE2_VER}.tar.gz"
    tar -xf "pcre2-${PCRE2_VER}.tar.gz"
    mv "pcre2-${PCRE2_VER}" pcre2
    rm "pcre2-${PCRE2_VER}.tar.gz"
fi

# 7. XZ (LZMA)
if [ ! -d "xz" ]; then
    XZ_VER="5.4.4"
    wget -q "https://github.com/tukaani-project/xz/releases/download/v${XZ_VER}/xz-${XZ_VER}.tar.gz"
    tar -xf "xz-${XZ_VER}.tar.gz"
    mv "xz-${XZ_VER}" xz
    rm "xz-${XZ_VER}.tar.gz"
fi

# 8. LibUV
if [ ! -d "libuv" ]; then
    wget -q "https://dist.libuv.org/dist/v1.48.0/libuv-v1.48.0.tar.gz"
    tar -xf libuv-v1.48.0.tar.gz
    mv libuv-v1.48.0 libuv
    rm libuv-v1.48.0.tar.gz
fi

# 9. cJSON
if [ ! -d "cjson" ]; then
    wget -q "https://github.com/DaveGamble/cJSON/archive/refs/tags/v1.7.18.tar.gz" -O cjson.tar.gz
    tar -xf cjson.tar.gz
    mv cJSON-1.7.18 cjson
    rm cjson.tar.gz
fi

echo "✔️ Sources downloaded."

build_lib() {
    local lib_name=$1
    local target=$2

    # Map the internal Zig target names to our clean output folder names
    local folder_target="x86_64-linux"
    if [[ "$target" == *"windows"* ]]; then folder_target="x86_64-windows-gnu"; fi

    local out_dir="$BIN_DIR/$folder_target/$lib_name"
    mkdir -p "$out_dir"

    # Skip if already built (checks for .so, .dll, .dylib)
    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.dylib 1> /dev/null 2>&1; then
        echo "✔️ $lib_name for $target already exists. Skipping."
        return
    fi

    echo "🔨 Building $lib_name for $target..."
    pushd "$EXT_DIR/$lib_name" > /dev/null

    # Use a Bash Array for safe argument passing
    local cmake_flags=(
        "-DCMAKE_C_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_CXX_COMPILER=zig;c++;-target;$target"
        "-DCMAKE_ASM_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_BUILD_TYPE=MinSizeRel"
    )
    local extra_c_flags=""
    
    if [[ "$target" == *"windows"* ]]; then
        cmake_flags+=("-DCMAKE_SYSTEM_NAME=Windows")
        extra_c_flags="-lws2_32" 
    else
        extra_c_flags="-pthread" 
    fi

    case $lib_name in
        mongoose)
            local out_file="libmongoose.so"
            if [[ "$target" == *"windows"* ]]; then out_file="mongoose.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" mongoose.c -D_FILE_OFFSET_BITS=64 -O3 $extra_c_flags
            ;;
        yyjson)
            local out_file="libyyjson.so"
            if [[ "$target" == *"windows"* ]]; then out_file="yyjson.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" yyjson.c -I. -O3
            ;;
        sqlite3)
            local out_file="libsqlite3.so"
            if [[ "$target" == *"windows"* ]]; then out_file="sqlite3.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" sqlite3.c -O3 $extra_c_flags
            ;;
        cjson)
            local out_file="libcjson.so"
            if [[ "$target" == *"windows"* ]]; then out_file="cjson.dll"; fi
            zig cc -target "$target" -shared -o "$out_dir/$out_file" cJSON.c -fPIC -O3
            ;;
        zlib)
            if [[ "$target" == *"windows"* ]]; then
                sed -i 's/if(MINGW)/if(FALSE)/g' CMakeLists.txt || true
            fi
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}"
            cmake --build build-$target --target zlib --parallel "$(nproc)"
            # FIX: More inclusive wildcard for Windows DLLs
            find build-$target -name "libz.so*" -o -name "*z*.dll" -o -name "libz.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
        zstd)
            rm -rf build-$target
            cmake -B build-$target -S build/cmake "${cmake_flags[@]}" -DZSTD_BUILD_STATIC=OFF -DZSTD_BUILD_PROGRAMS=OFF -DZSTD_ASSEMBLY_DISABLE=ON
            cmake --build build-$target --parallel "$(nproc)"
            # FIX: More inclusive wildcard for Windows DLLs
            find build-$target -name "libzstd.so*" -o -name "*zstd*.dll" -o -name "libzstd.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
        xz)
            if [[ "$target" == *"windows"* ]]; then
                perl -0777 -i -pe 's/target_sources\([^)]*\.rc[^)]*\)//g' CMakeLists.txt || true
                perl -0777 -i -pe 's/set_target_properties\([^)]*\.rc[^)]*\)//g' CMakeLists.txt || true
            fi
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" -DBUILD_SHARED_LIBS=ON
            cmake --build build-$target --target liblzma --parallel "$(nproc)"
            # FIX: More inclusive wildcard for Windows DLLs
            find build-$target -name "liblzma.so*" -o -name "*lzma*.dll" -o -name "liblzma.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
        pcre2)
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" -DBUILD_SHARED_LIBS=ON -DPCRE2_BUILD_PCRE2_8=ON -DPCRE2_BUILD_PCRE2_16=OFF -DPCRE2_BUILD_PCRE2_32=OFF -DPCRE2_BUILD_TESTS=OFF -DPCRE2_BUILD_PCRE2GREP=OFF
            cmake --build build-$target --parallel "$(nproc)"
            find build-$target -name "*pcre2-8.so*" -o -name "*pcre2-8*.dll" -o -name "*pcre2-8.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
        libuv)
            rm -rf build-$target
            cmake -B build-$target "${cmake_flags[@]}" -DBUILD_TESTING=OFF
            cmake --build build-$target --parallel "$(nproc)"
            find build-$target -name "libuv.so*" -o -name "*uv*.dll" -o -name "libuv.dylib*" | xargs -I {} cp {} "$out_dir/"
            ;;
    esac
    
    if [[ "$target" == *"linux"* ]]; then
        cd "$out_dir"
        for f in *.so.*; do
            [ -e "$f" ] || continue
            base="${f%%.so*}.so"
            ln -sf "$f" "$base"
        done
    fi

    popd > /dev/null
}

LIBS=("mongoose" "yyjson" "sqlite3" "cjson" "zlib" "zstd" "xz" "pcre2" "libuv") # mbedtls is external

for target in "${TARGETS[@]}"; do
    for lib in "${LIBS[@]}"; do
        build_lib "$lib" "$target"
    done
done

echo "✅ All standard shared libraries built successfully and placed in $BIN_DIR"