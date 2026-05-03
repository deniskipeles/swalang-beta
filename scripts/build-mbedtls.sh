#!/bin/bash
set -euo pipefail

# =============================================================================
# build-mbedtls.sh — Cross-compile MbedTLS for Linux and Windows via Zig
#
# Self-healing behaviors:
#   - Auto-restores any patched submodule files before AND after patching
#   - Cleans stale/dirty build dirs on CMake configure failure
#   - Works in GitHub Actions (no interactive assumptions)
# =============================================================================

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXT_DIR="$PROJECT_ROOT/.extensions"
BIN_DIR="$PROJECT_ROOT/bin"
MBEDTLS_DIR="$EXT_DIR/mbedtls"
export TFPSA_CMAKE="$MBEDTLS_DIR/tf-psa-crypto/core/CMakeLists.txt"

TARGETS=("x86_64-linux-gnu.2.17" "x86_64-windows-gnu")

# -----------------------------------------------------------------------------
# Restore the tf-psa-crypto CMakeLists to a known-clean state.
# Called at script start, on EXIT trap, and after each Windows build.
# Safe to call even if the file was never patched.
# -----------------------------------------------------------------------------
restore_tfpsa_cmake() {
    if [ -f "$TFPSA_CMAKE" ]; then
        git -C "$MBEDTLS_DIR/tf-psa-crypto" checkout core/CMakeLists.txt 2>/dev/null || true
    fi
}

# Always restore on any exit (success, error, or signal) so CI never
# leaves a dirty submodule that poisons the next job or matrix leg.
trap restore_tfpsa_cmake EXIT

mkdir -p "$EXT_DIR"

echo "================================================="
echo "🛡️  Building Isolated MbedTLS FFI Dependency"
echo "================================================="

# -----------------------------------------------------------------------------
# 1. Clone or refresh MbedTLS
# -----------------------------------------------------------------------------
if [ ! -d "$MBEDTLS_DIR" ]; then
    echo "📥 Cloning MbedTLS and submodules..."
    git clone --depth 1 --recurse-submodules \
        https://github.com/Mbed-TLS/mbedtls.git "$MBEDTLS_DIR"
else
    echo "✔️  MbedTLS source already present."
    # Ensure submodules are initialised (e.g. after a shallow CI cache restore)
    git -C "$MBEDTLS_DIR" submodule update --init --recursive 2>/dev/null || true
    # Unconditionally restore any leftover patch from a previous failed run
    restore_tfpsa_cmake
fi

# -----------------------------------------------------------------------------
# 2. Build for each target
# -----------------------------------------------------------------------------
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

    # Skip only if we already have real output artifacts
    if ls "$out_dir"/*.so "$out_dir"/*.dll "$out_dir"/*.a "$out_dir"/*.dylib \
           1>/dev/null 2>&1; then
        echo "✔️  mbedtls for $target already built. Skipping."
        continue
    fi

    echo ""
    echo "🔨 Compiling mbedtls for $target..."

    # -------------------------------------------------------------------------
    # Windows-only: patch tfpsacrypto SHARED → STATIC before CMake sees it.
    #
    # Root cause: zig's lld injects compiler_rt.lib (which contains __chkstk_ms)
    # into every shared library it links. With three DLLs in a link chain
    # (libmbedcrypto → libmbedx509 → libmbedtls), each carries its own copy and
    # lld hard-errors on the duplicate. Building everything static means the
    # symbol is resolved exactly once at the final link step.
    # -------------------------------------------------------------------------
    if $is_windows; then
        # Guard: ensure the file is clean before we touch it
        restore_tfpsa_cmake

        echo "🩹 Patching tf-psa-crypto: SHARED → STATIC for $target..."

        # TFPSA_CMAKE is exported above so Python can read it from os.environ
        python3 - <<'PYEOF'
import sys, os, pathlib

path = pathlib.Path(os.environ["TFPSA_CMAKE"])

original = path.read_text()
if "add_library(${tfpsacrypto_target} SHARED" not in original:
    print("ℹ️  Pattern not found — already patched or upstream changed. Skipping.")
    sys.exit(0)

patched = original.replace(
    "add_library(${tfpsacrypto_target} SHARED",
    "add_library(${tfpsacrypto_target} STATIC",
)
path.write_text(patched)
print("✅ Patch applied.")
PYEOF

        echo "🔍 Verifying patch..."
        grep -n "add_library.*tfpsacrypto" "$TFPSA_CMAKE"
    fi

    # -------------------------------------------------------------------------
    # CMake flags
    # -------------------------------------------------------------------------
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
            # Suppress __chkstk_ms at the compiler level too
            "-DCMAKE_C_FLAGS=-mno-stack-arg-probe -fno-stack-check"
            # All static — no DLL chain, no compiler_rt collision
            "-DUSE_SHARED_MBEDTLS_LIBRARY=OFF"
            "-DUSE_STATIC_MBEDTLS_LIBRARY=ON"
        )
    else
        cmake_flags+=(
            "-DUSE_SHARED_MBEDTLS_LIBRARY=ON"
        )
    fi

    # Clean any stale build dir (e.g. from a previously failed configure)
    rm -rf "$build_dir"

    # Run cmake from the mbedtls source root
    pushd "$MBEDTLS_DIR" > /dev/null

        cmake -B "$build_dir" "${cmake_flags[@]}"

        # No --parallel on Windows: avoids residual .a write races between
        # tfpsacrypto and tfpsacrypto_static targets writing to the same path
        if $is_windows; then
            cmake --build "$build_dir"
        else
            cmake --build "$build_dir" --parallel "$(nproc)"
        fi

    popd > /dev/null

    # Restore patch immediately after a successful Windows build
    # (the EXIT trap is the safety net for failures)
    if $is_windows; then
        echo "🔄 Restoring tf-psa-crypto/core/CMakeLists.txt..."
        restore_tfpsa_cmake
    fi

    # -------------------------------------------------------------------------
    # Copy artifacts to output directory
    # -------------------------------------------------------------------------
    if $is_windows; then
        # Static libs only — DLLs were never built
        find "$build_dir" \( -name "libmbed*.a" -o -name "libtfpsa*.a" \) \
            -exec cp {} "$out_dir/" \;
    else
        # Shared libs + versioned symlinks for Linux
        find "$build_dir" \( -name "libmbed*.so*" -o -name "libmbed*.dylib*" \) \
            -exec cp -P {} "$out_dir/" \;
        # Ensure bare .so symlinks exist (e.g. libmbedtls.so → libmbedtls.so.3.6.1)
        pushd "$out_dir" > /dev/null
        for f in *.so.*; do
            [ -e "$f" ] || continue
            base="${f%%.so*}.so"
            [ -L "$base" ] || ln -sf "$f" "$base"
        done
        popd > /dev/null
    fi

    echo "✅ Success for $target"
    echo "   └─ artifacts: $(ls "$out_dir" | tr '\n' ' ')"
done

echo ""
echo "🎉 Isolated MbedTLS build complete!"