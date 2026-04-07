#!/usr/bin/env bash
# Build libffi for Android (NDK) and/or Windows (mingw-w64).
# Usage:
#   bash build-libffi.sh --target android       # interactive arch selection for Android
#   bash build-libffi.sh --target android --arch arm64 --api 28
#   bash build-libffi.sh --target windows
#   bash build-libffi.sh --target all
#   bash build-libffi.sh --help
set -euo pipefail

LIBFFI_VERSION="${LIBFFI_VERSION:-3.4.6}"
WORKDIR="${WORKDIR:-$(pwd)/libffi-build}"
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK:-$HOME/android-ndk-r26d}}"
API_LEVEL_DEFAULT="${ANDROID_API_LEVEL:-28}"

print_usage() {
    cat <<EOF
build-libffi.sh - build libffi for Android and/or Windows

Options:
  --target <android|windows|all>    Which target to build (default: android)
  --arch <arm64|arm|x86|x86_64>     Android arch (optional; interactive if omitted for android)
  --api <API_LEVEL>                 Android API level (default: ${API_LEVEL_DEFAULT})
  --libffi-version <ver>            libffi version (default: ${LIBFFI_VERSION})
  --workdir <path>                  base workdir (default: ${WORKDIR})
  --ndk <path>                      path to Android NDK (default: ${NDK})
  --help                            show this help
Examples:
  bash build-libffi.sh --target android --arch arm64 --api 28
  bash build-libffi.sh --target windows
  bash build-libffi.sh --target all
EOF
}

# Simple arg parse
TARGET="android"
ARCH_ARG=""
API_LEVEL="$API_LEVEL_DEFAULT"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --target) TARGET="$2"; shift 2;;
        --arch) ARCH_ARG="$2"; shift 2;;
        --api) API_LEVEL="$2"; shift 2;;
        --libffi-version) LIBFFI_VERSION="$2"; shift 2;;
        --workdir) WORKDIR="$2"; shift 2;;
        --ndk) NDK="$2"; shift 2;;
        --help) print_usage; exit 0;;
        *) echo "Unknown arg: $1"; print_usage; exit 1;;
    esac
done

mkdir -p "$WORKDIR"
cd "$WORKDIR"

TARBALL="libffi-${LIBFFI_VERSION}.tar.gz"
SRC_DIR="$WORKDIR/libffi-$LIBFFI_VERSION"

download_libffi() {
    if [ ! -f "$TARBALL" ]; then
        echo "Downloading libffi v${LIBFFI_VERSION}..."
        wget -q --show-progress "https://github.com/libffi/libffi/releases/download/v${LIBFFI_VERSION}/libffi-${LIBFFI_VERSION}.tar.gz" -O "$TARBALL"
    else
        echo "Using existing $TARBALL"
    fi

    if [ ! -d "$SRC_DIR" ]; then
        tar xf "$TARBALL"
    else
        echo "Source already extracted at $SRC_DIR"
    fi
}

# Build helper for Windows (mingw)
build_windows() {
    echo ""
    echo "=== Building libffi for Windows (x86_64-w64-mingw32) ==="
    # Check for mingw-w64 toolchain available (x86_64-w64-mingw32-gcc)
    if ! command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1 && ! command -v x86_64-w64-mingw32-clang >/dev/null 2>&1; then
        echo "Warning: mingw-w64 cross-compiler not found (x86_64-w64-mingw32-gcc). Install mingw-w64."
        echo "On Ubuntu: sudo apt install mingw-w64"
        # We'll still attempt configure; it will fail if compiler missing.
    fi

    local builddir="$WORKDIR/build-windows-x86_64"
    local installdir="$WORKDIR/install-windows-x86_64"
    rm -rf "$builddir" "$installdir"
    mkdir -p "$builddir" "$installdir"
    pushd "$builddir" >/dev/null

    CC="${MINGW_CC:-x86_64-w64-mingw32-gcc}"
    export CC

    echo "Configuring for host=x86_64-w64-mingw32 prefix=${installdir} ..."
    "$SRC_DIR/configure" --host=x86_64-w64-mingw32 --prefix="$installdir" --disable-shared --enable-static
    make -j"$(nproc)"
    make install
    popd >/dev/null

    echo "Windows build complete:"
    echo "  Headers: $installdir/include"
    echo "  Static lib: $installdir/lib/libffi.a"
}

# Build helper for Android
build_android_for_arch() {
    local ARCH="$1"   # arm64 arm x86 x86_64
    local TOOLCHAIN_PREFIX
    case "$ARCH" in
        arm64) TOOLCHAIN_PREFIX="aarch64-linux-android"; GOARCH="arm64";;
        arm) TOOLCHAIN_PREFIX="armv7a-linux-androideabi"; GOARCH="arm";;
        x86) TOOLCHAIN_PREFIX="i686-linux-android"; GOARCH="386";;
        x86_64) TOOLCHAIN_PREFIX="x86_64-linux-android"; GOARCH="amd64";;
        *) echo "Unsupported arch: $ARCH"; return 1;;
    esac

    if [ ! -d "$NDK" ]; then
        echo "NDK not found at $NDK. Attempting to download android-ndk-r26d to $NDK ..."
        # attempt download into parent folder of $NDK
        PARENT="$(dirname "$NDK")"
        mkdir -p "$PARENT"
        ZIP="$PARENT/android-ndk-r26d-linux.zip"
        wget -q --show-progress https://dl.google.com/android/repository/android-ndk-r26d-linux.zip -O "$ZIP"
        unzip -q "$ZIP" -d "$PARENT"
        rm -f "$ZIP"
        if [ ! -d "$NDK" ]; then
            # some NDK zip extracts to android-ndk-r26d (no change) or android-ndk-r26d (should match)
            echo "Downloaded but expected NDK at $NDK not found. Please set --ndk to the correct path."
            return 1
        fi
    fi

    local TOOLCHAIN_BIN="$NDK/toolchains/llvm/prebuilt/linux-x86_64/bin"
    if [ ! -d "$TOOLCHAIN_BIN" ]; then
        echo "Toolchain bin not found at $TOOLCHAIN_BIN"
        return 1
    fi

    local CC="$TOOLCHAIN_BIN/${TOOLCHAIN_PREFIX}${API_LEVEL}-clang"
    local CXX="$TOOLCHAIN_BIN/${TOOLCHAIN_PREFIX}${API_LEVEL}-clang++"
    local AR="$TOOLCHAIN_BIN/llvm-ar"
    local RANLIB="$TOOLCHAIN_BIN/llvm-ranlib"
    local STRIP="$TOOLCHAIN_BIN/llvm-strip"
    local LD="$TOOLCHAIN_BIN/ld.lld"

    if [ ! -x "$CC" ]; then
        echo "Warning: expected clang not found at $CC"
    fi

    local builddir="$WORKDIR/build-android-$ARCH"
    local installdir="$WORKDIR/install-android-$ARCH"
    rm -rf "$builddir" "$installdir"
    mkdir -p "$builddir" "$installdir"
    pushd "$builddir" >/dev/null

    echo "Configuring libffi for Android ($ARCH) ..."
    export CC CXX AR RANLIB STRIP LD
    export CFLAGS="--sysroot=$NDK/toolchains/llvm/prebuilt/linux-x86_64/sysroot"
    export LDFLAGS="--sysroot=$NDK/toolchains/llvm/prebuilt/linux-x86_64/sysroot"

    # On some versions configure expects --host like aarch64-linux-android
    "$SRC_DIR/configure" \
        --host="${TOOLCHAIN_PREFIX}" \
        --prefix="$installdir" \
        --disable-shared \
        --enable-static

    make -j"$(nproc)"
    make install

    popd >/dev/null

    echo "Android $ARCH build complete:"
    echo "  Headers: $installdir/include"
    echo "  Static lib: $installdir/lib/libffi.a"
    echo ""
    echo "To use with Go (example):"
    echo "  export CGO_CFLAGS=\"-I$installdir/include\""
    echo "  export CGO_LDFLAGS=\"-L$installdir/lib -lffi\""
    echo "  export CC=$CC"
}

# MAIN
download_libffi

if [[ "$TARGET" == "android" || "$TARGET" == "all" ]]; then
    # determine arches to build
    if [[ -n "$ARCH_ARG" ]]; then
        ARCHS_TO_BUILD=("$ARCH_ARG")
    else
        # interactive selection if running in terminal and no arch given
        if [ -t 0 ]; then
            echo "Select Android architecture to build for:"
            PS3="#? "
            select ARCH in "arm64" "x86" "arm" "x86_64" "all"; do
                if [[ "$ARCH" == "all" ]]; then
                    ARCHS_TO_BUILD=(arm64 arm x86 x86_64)
                    break
                elif [[ -n "$ARCH" ]]; then
                    ARCHS_TO_BUILD=("$ARCH")
                    break
                fi
            done
        else
            echo "No --arch provided and not interactive; defaulting to arm64"
            ARCHS_TO_BUILD=(arm64)
        fi
    fi

    for A in "${ARCHS_TO_BUILD[@]}"; do
        build_android_for_arch "$A"
    done
fi

if [[ "$TARGET" == "windows" || "$TARGET" == "all" ]]; then
    build_windows
fi

echo ""
echo "All requested builds finished. Artifacts are in: $WORKDIR"
echo "Example install dirs:"
ls -1 "$WORKDIR" | sed -n '1,200p' || true