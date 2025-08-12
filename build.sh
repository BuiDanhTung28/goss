#!/bin/bash
# build.sh - Cross-platform build script for macOS and Linux

set -e

OS="$(uname)"
echo "==> Detected OS: $OS"

# Check dependencies based on OS
if [[ "$OS" == "Linux" ]]; then
  if ! (command -v cmake &> /dev/null && command -v make &> /dev/null && command -v g++ &> /dev/null); then
    echo "Error: cmake, make, g++ required."
    echo "Install with: sudo apt-get install build-essential cmake"
    exit 1
  fi
elif [[ "$OS" == "Darwin" ]]; then
  if ! xcode-select -p &>/dev/null; then
    echo "Error: Xcode Command Line Tools required."
    exit 1
  fi
  if ! command -v brew &> /dev/null; then
    echo "Error: Homebrew required."
    exit 1
  fi
  if ! (brew ls --versions cmake &> /dev/null && brew ls --versions libomp &> /dev/null); then
    echo "Error: cmake and libomp required."
    echo "Install with: brew install cmake libomp"
    exit 1
  fi
  export CMAKE_PREFIX_PATH=$(brew --prefix libomp):$CMAKE_PREFIX_PATH
  echo "Added libomp to CMAKE_PREFIX_PATH"
else
  echo "Unsupported operating system: $OS"
  echo "Supported platforms: macOS, Linux"
  exit 1
fi

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
FAISS_SOURCE_DIR="$ROOT_DIR/faiss_source"

# Determine platform and architecture
if [[ "$OS" == "Darwin" ]]; then
    ARCH=$(uname -m)
    if [[ "$ARCH" == "arm64" ]]; then
        PLATFORM="darwin"
        ARCH_NAME="darwin_arm64"
        CMAKE_FLAGS="-DBLA_VENDOR=Apple"
    else
        PLATFORM="darwin"
        ARCH_NAME="darwin_x64"
        CMAKE_FLAGS="-DBLA_VENDOR=Apple"
    fi
elif [[ "$OS" == "Linux" ]]; then
    ARCH=$(uname -m)
    if [[ "$ARCH" == "x86_64" ]]; then
        PLATFORM="linux"
        ARCH_NAME="linux_x64"
        CMAKE_FLAGS="-DBLA_VENDOR=OpenBLAS"
    elif [[ "$ARCH" == "aarch64" ]]; then
        PLATFORM="linux"
        ARCH_NAME="linux_arm64"
        CMAKE_FLAGS="-DBLA_VENDOR=OpenBLAS"
    else
        echo "Unsupported Linux architecture: $ARCH"
        exit 1
    fi
fi

FAISS_LIB_DIR="$ROOT_DIR/internal/lib/$ARCH_NAME"
FAISS_STATIC_LIB="$FAISS_LIB_DIR/libfaiss.a"
FAISS_C_STATIC_LIB="$FAISS_LIB_DIR/libfaiss_c.a"

# Check if libraries already exist
if [ -f "$FAISS_STATIC_LIB" ] && [ -f "$FAISS_C_STATIC_LIB" ]; then
  echo "FAISS static libraries for $PLATFORM ($ARCH_NAME) already exist. Skipping build."
  exit 0
fi

echo "Building FAISS static libraries for $PLATFORM ($ARCH_NAME)..."

FAISS_BUILD_DIR="$ROOT_DIR/internal/build/$ARCH_NAME"

mkdir -p "$FAISS_BUILD_DIR"

cmake -S "$FAISS_SOURCE_DIR" -B "$FAISS_BUILD_DIR" \
    -DFAISS_ENABLE_GPU=OFF \
    -DFAISS_ENABLE_PYTHON=OFF \
    -DBUILD_SHARED_LIBS=OFF \
    -DFAISS_ENABLE_C_API=ON \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
    $CMAKE_FLAGS

# Build core library
make -C "$FAISS_BUILD_DIR" -j faiss

# Build C API static library
make -C "$FAISS_BUILD_DIR" -j faiss_c

mkdir -p "$FAISS_LIB_DIR"

cp "$FAISS_BUILD_DIR/faiss/libfaiss.a" "$FAISS_STATIC_LIB"
cp "$FAISS_BUILD_DIR/c_api/libfaiss_c.a" "$FAISS_C_STATIC_LIB"

echo "FAISS static libraries for $PLATFORM ($ARCH_NAME) built successfully."
echo "Libraries are available in: $FAISS_LIB_DIR"
