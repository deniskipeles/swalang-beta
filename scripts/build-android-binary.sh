#!/usr/bin/env bash
# ./scripts/build-android-binary.sh
# This script builds temporary binaries and prints ONLY the architecture name to stdout on success.
# All user-facing messages are redirected to stderr.
set -e

# --- NDK Setup ---
NDK_DIR="$HOME/android-ndk-r26d"
if [ -z "$ANDROID_NDK_HOME" ] && [ -z "$ANDROID_NDK" ]; then
    if [ ! -d "$NDK_DIR" ]; then
        echo "Android NDK not found at $NDK_DIR. Downloading..." >&2
        wget -q --show-progress https://dl.google.com/android/repository/android-ndk-r26d-linux.zip -O "$HOME/android-ndk-r26d-linux.zip"
        unzip -q "$HOME/android-ndk-r26d-linux.zip" -d "$HOME"
        rm "$HOME/android-ndk-r26d-linux.zip"
    fi
    export ANDROID_NDK="$NDK_DIR"
else
    NDK_DIR="${ANDROID_NDK_HOME:-$ANDROID_NDK}"
fi
echo "Using Android NDK at $NDK_DIR." >&2

# --- Interactive Selections ---
declare -A ARCHS
ARCHS=(
    [arm64]="aarch64-linux-android"
    [arm]="armv7a-linux-androideabi"
    [x86]="i686-linux-android"
    [x86_64]="x86_64-linux-android"
)

echo "Select Android architecture for Termux:" >&2
select ARCH in "${!ARCHS[@]}"; do
    if [[ -n "$ARCH" ]]; then
        TARGET_ARCH=$ARCH
        TOOLCHAIN=${ARCHS[$ARCH]}
        break
    else
        exit 1
    fi
done

# The 'read' prompt automatically goes to stderr, which is correct.
read -p "Enter your target Android API level (e.g., 24, 28, 30): " API_LEVEL

declare -A LANGUAGES
LANGUAGES=( [swahili]="sw" [english]="en" )

echo "Select targeted Language for Termux Binary:" >&2
select LANG in "${!LANGUAGES[@]}"; do
    if [[ -n "$LANG" ]]; then
        TARGET_TAGS=${LANGUAGES[$LANG]}
        break
    else
        exit 1
    fi
done

# --- Build Environment Setup ---
PREBUILT="$NDK_DIR/toolchains/llvm/prebuilt/linux-x86_64/bin"
export GOOS=android
export GOARCH=$TARGET_ARCH
export CGO_ENABLED=1
export CC="$PREBUILT/${TOOLCHAIN}${API_LEVEL}-clang"
export CXX="$PREBUILT/${TOOLCHAIN}${API_LEVEL}-clang++"
export CGO_CFLAGS="-I/home/zeus/content/install-android-${GOARCH}/include"
export CGO_LDFLAGS="-L/home/zeus/content/install-android-${GOARCH}/lib -lffi"

# --- Build Commands ---
TEMP_DIR="builds/android-temp"
TEMP_NORMAL_DIR="${TEMP_DIR}/normal"
TEMP_MINI_DIR="${TEMP_DIR}/mini"
mkdir -p "${TEMP_NORMAL_DIR}" "${TEMP_MINI_DIR}"

echo "Building Android binaries for ${GOARCH}..." >&2
go build -tags ${TARGET_TAGS} -o "${TEMP_NORMAL_DIR}/swalang-${GOARCH}" ./cmd/interpreter/main.go
go build -tags ${TARGET_TAGS} -ldflags="-s -w" -o "${TEMP_MINI_DIR}/swalang-${GOARCH}" ./cmd/interpreter/main.go

# --- Communication ---
# This is the ONLY output on stdout. This is the script's "return value".
echo "$TARGET_ARCH"