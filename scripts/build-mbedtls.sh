#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"
MBEDTLS_DIR="$EXT_DIR/mbedtls"
export TFPSA_CMAKE="$MBEDTLS_DIR/tf-psa-crypto/core/CMakeLists.txt"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

restore_tfpsa_cmake() {
    if [ -f "$TFPSA_CMAKE" ]; then
        git -C "$MBEDTLS_DIR/tf-psa-crypto" checkout core/CMakeLists.txt 2>/dev/null || true
    fi
}

trap restore_tfpsa_cmake EXIT

mkdir -p "$EXT_DIR"

echo "================================================="
echo "🛡️  Building Isolated MbedTLS FFI Dependency"
echo "================================================="

if [ ! -d "$MBEDTLS_DIR" ]; then
    echo "📥 Cloning MbedTLS and submodules..."
    git clone --depth 1 --recurse-submodules \
        https://github.com/Mbed-TLS/mbedtls.git "$MBEDTLS_DIR"
else
    echo "✔️  MbedTLS source already present."
    git -C "$MBEDTLS_DIR" submodule update --init --recursive 2>/dev/null || true
    restore_tfpsa_cmake
fi

for target in "${TARGETS[@]}"; do
    is_windows=false
    folder_target="x86_64-linux"
    if [[ "$target" == *"windows"* ]]; then
        is_windows=true
        folder_target="x86_64-windows-gnu"
    fi

    out_dir="$BIN_DIR/$folder_target/mbedtls"
    mkdir -p "$out_dir"
    build_dir="$MBEDTLS_DIR/build-$target"

    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.a "$out_dir"/*.dylib \
           1>/dev/null 2>&1; then
        echo "✔️  mbedtls for $target already built. Skipping."
        continue
    fi

    echo ""
    echo "🔨 Compiling mbedtls for $target..."

    if $is_windows; then
        restore_tfpsa_cmake

        echo "🩹 Patching tf-psa-crypto: SHARED → STATIC for $target..."
        python3 - <<'PYEOF'
import sys, os, pathlib

path = pathlib.Path(os.environ["TFPSA_CMAKE"])
original = path.read_text()

if "add_library(${tfpsacrypto_target} SHARED" not in original:
    print("ℹ️  Already patched or upstream changed. Skipping.")
    sys.exit(0)

path.write_text(original.replace(
    "add_library(${tfpsacrypto_target} SHARED",
    "add_library(${tfpsacrypto_target} STATIC",
))
print("✅ Patch applied.")
PYEOF

        echo "🔍 Verifying patch..."
        grep -n "add_library.*tfpsacrypto" "$TFPSA_CMAKE"
    fi

    cmake_flags=(
        "-DCMAKE_C_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_CXX_COMPILER=zig;c++;-target;$target"
        "-DCMAKE_ASM_COMPILER=zig;cc;-target;$target"
        "-DCMAKE_BUILD_TYPE=MinSizeRel"
        "-DENABLE_TESTING=OFF"
        "-DENABLE_PROGRAMS=OFF"
        "-DMBEDTLS_FATAL_WARNINGS=OFF"
    )

    if $is_windows; then
        cmake_flags+=(
            "-DCMAKE_SYSTEM_NAME=Windows"
            "-DCMAKE_C_FLAGS=-mno-stack-arg-probe -fno-stack-check"
            "-DUSE_SHARED_MBEDTLS_LIBRARY=OFF"
            "-DUSE_STATIC_MBEDTLS_LIBRARY=ON"
        )
    else
        cmake_flags+=(
            "-DUSE_SHARED_MBEDTLS_LIBRARY=ON"
            "-DUSE_STATIC_MBEDTLS_LIBRARY=OFF"
        )
    fi

    rm -rf "$build_dir"

    pushd "$MBEDTLS_DIR" > /dev/null
        cmake -B "$build_dir" "${cmake_flags[@]}"
        cmake --build "$build_dir" --parallel "$(nproc)"
    popd > /dev/null

    if $is_windows; then
        echo "🔄 Restoring tf-psa-crypto/core/CMakeLists.txt..."
        restore_tfpsa_cmake

        # libtfpsacrypto.a excluded — symbols already baked into libmbedcrypto.a
        declare -A seen_libs
        static_libs=()
        while IFS= read -r lib; do
            base="$(basename "$lib")"
            if [[ -z "${seen_libs[$base]+x}" ]]; then
                seen_libs[$base]=1
                static_libs+=("$lib")
            fi
        done < <(find "$build_dir" -name "libmbed*.a" | sort)
        unset seen_libs

        echo "🔗 Merging static libs into a single libmbedtls.dll..."
        echo "   libs: ${static_libs[*]}"

        zig cc -target "$target" -shared \
            -nodefaultlibs \
            -Wl,--whole-archive \
            "${static_libs[@]}" \
            -Wl,--no-whole-archive \
            -lkernel32 -lntdll -lws2_32 -lbcrypt \
            -o "$out_dir/libmbedtls.dll"

        echo "✅ libmbedtls.dll created."
    else
        # Copy only real files (not symlinks) — we rebuild symlinks cleanly below.
        # Include libtfpsa*.so* because libmbedcrypto.so.X.Y.Z symlinks into it.
        find "$build_dir" \( \
            -name "libmbed*.so*" -o \
            -name "libmbed*.dylib*" -o \
            -name "libtfpsa*.so*" \
        \) -not -type l -exec cp {} "$out_dir/" \;

        pushd "$out_dir" > /dev/null

        # Pass 1: bare .so → real versioned file
        for real in *.so.*.*; do
            [ -f "$real" ] || continue
            bare="${real%%.so*}.so"
            [ -L "$bare" ] || ln -sf "$real" "$bare"
        done

        # Pass 2: intermediate SONAME symlink (e.g. libmbedcrypto.so.18 → libmbedcrypto.so.4.1.0)
        # Read the SONAME embedded in each .so file and create that symlink if missing
        for real in *.so.*.*; do
            [ -f "$real" ] || continue
            soname=$(readelf -d "$real" 2>/dev/null \
                | awk '/SONAME/ { gsub(/.*\[|\].*/, ""); print; exit }')
            if [[ -n "$soname" && "$soname" != "$real" && ! -e "$soname" ]]; then
                ln -sf "$real" "$soname"
            fi
        done

        popd > /dev/null
    fi

    echo "✅ Success for $target"
    echo "   └─ artifacts: $(ls "$out_dir" | tr '\n' ' ')"
done

echo ""
echo "🎉 Isolated MbedTLS build complete!"