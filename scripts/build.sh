#!/bin/bash
# ./scripts/build.sh
# Create the swalang binaries for multiple platforms, handling CGO and Android builds.

set -euo pipefail

# --- Configuration ---
OUTPUT_BASE_NAME="swalang"
PACKAGE_PATH="cmd/interpreter/main.go"

# --- Determine Project Root & Change Directory ---
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "Project root identified as: ${PROJECT_ROOT}"
cd "${PROJECT_ROOT}"

OS_ARRAY=("linux" "windows" "android")
ARCH_ARRAY=("amd64")

# --- Pre-build Steps ---
echo "Cleaning up old build directories..."
rm -rf builds/

echo "Running go mod tidy..."
go mod tidy

if command -v upx >/dev/null 2>&1; then
  USE_UPX=true
  echo "UPX found. Binaries will be compressed where possible."
else
  USE_UPX=false
  echo "Warning: UPX not found. Skipping final compression step."
fi

if [[ " ${OS_ARRAY[*]} " =~ " windows " ]]; then
    echo "--------------------------------------------------"
    echo "Windows target detected. Running prerequisite scripts..."
    bash scripts/install-windows-libffi.sh
fi

echo "--------------------------------------------------"
echo "Starting cross-platform builds..."

# --- Main Build Loop ---
for OS in "${OS_ARRAY[@]}"; do
  # The Android script handles its own architecture loop/selection.
  if [ "$OS" = "android" ]; then
    echo "-> Android build selected. Handing off to dedicated script..."
    
    ANDROID_ARCH=$(bash scripts/build-android-binary.sh || echo "skip")
    
    if [ "$ANDROID_ARCH" = "skip" ]; then
        echo "--> Android build skipped by user."
        continue
    fi
    
    echo "--> Android build complete for arch: ${ANDROID_ARCH}. Organizing files..."

    BASE_DIR="builds/android/${ANDROID_ARCH}"
    MINI_DIR="${BASE_DIR}/mini"
    UPX_DIR="${BASE_DIR}/upx"
    mkdir -p "${BASE_DIR}" "${MINI_DIR}" "${UPX_DIR}"

    TEMP_DIR="builds/android-temp"
    TEMP_NORMAL_FILE="${TEMP_DIR}/normal/swalang-${ANDROID_ARCH}"
    TEMP_MINI_FILE="${TEMP_DIR}/mini/swalang-${ANDROID_ARCH}"

    mv "${TEMP_NORMAL_FILE}" "${BASE_DIR}/swalang"
    mv "${TEMP_MINI_FILE}" "${MINI_DIR}/swalang"

    # FIX: UPX is known to be incompatible with modern NDK output.
    # We will copy the file but skip the compression step for Android.
    if [ "$USE_UPX" = true ]; then
      echo "  -> NOTE: Skipping UPX for Android due to toolchain incompatibility."
      cp "${MINI_DIR}/swalang" "${UPX_DIR}/swalang"
    fi

    rm -rf "${TEMP_DIR}"
    
    echo "Finished organizing files for android/${ANDROID_ARCH}."
    echo "--------------------------------------------------"
    continue
  fi

  for ARCH in "${ARCH_ARRAY[@]}"; do
    echo "Building for ${OS}/${ARCH}..."
    
    if [ "$OS" = "windows" ]; then
      export GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=/usr/bin/x86_64-w64-mingw32-gcc
    else
      unset CC
      export GOOS=$OS GOARCH=$ARCH CGO_ENABLED=1
    fi

    EXT=""
    if [ "$OS" = "windows" ]; then EXT=".exe"; fi

    BASE_DIR="builds/${OS}/${ARCH}"
    MINI_DIR="${BASE_DIR}/mini"
    UPX_DIR="${BASE_DIR}/upx"
    mkdir -p "${BASE_DIR}" "${MINI_DIR}" "${UPX_DIR}"
    
    OUTPUT_NORMAL="${BASE_DIR}/${OUTPUT_BASE_NAME}${EXT}"
    OUTPUT_MINI="${MINI_DIR}/${OUTPUT_BASE_NAME}${EXT}"
    OUTPUT_UPX="${UPX_DIR}/${OUTPUT_BASE_NAME}${EXT}"
    
    echo "  -> Building normal binary..."
    go build -tags="sw" -o "${OUTPUT_NORMAL}" "${PACKAGE_PATH}"
    echo "  -> Building stripped binary..."
    go build -tags="sw" -ldflags="-s -w" -o "${OUTPUT_MINI}" "${PACKAGE_PATH}"

    if [ "$USE_UPX" = true ]; then
      echo "  -> Compressing with UPX..."
      cp "${OUTPUT_MINI}" "${OUTPUT_UPX}"
      upx --best --ultra-brute "${OUTPUT_UPX}" > /dev/null
    fi

    unset GOOS GOARCH CGO_ENABLED
    if [ "$OS" = "windows" ]; then unset CC; fi

    echo "Finished build for ${OS}/${ARCH}."
    echo "--------------------------------------------------"
  done
done

echo "✅ Build process complete."
echo "   Final binaries are located in the builds/ directory."