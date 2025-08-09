#!/bin/bash
# build.sh - Cross-platform build script

set -e

OS="$(uname)"
echo "==> Detected OS: $OS"

if [[ "$OS" == "Linux" ]]; then
  if ! (command -v cmake &> /dev/null && command -v make &> /dev/null && command -v g++ &> /dev/null); then
    echo "Error: cmake, make, g++ required."
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
    exit 1
  fi
  export CMAKE_PREFIX_PATH=$(brew --prefix libomp):$CMAKE_PREFIX_PATH
  echo "Added libomp to CMAKE_PREFIX_PATH"
fi

ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
FAISS_LIB_DIR="$ROOT_DIR/internal/lib/darwin_arm64"
FAISS_STATIC_LIB="$FAISS_LIB_DIR/libfaiss.a"

if [ -f "$FAISS_STATIC_LIB" ]; then
  echo "Faiss static library already exists. Skipping build."
  exit 0
fi

echo "Building Faiss static library with C API enabled..."

FAISS_SOURCE_DIR="$ROOT_DIR/faiss_source"
FAISS_BUILD_DIR="$ROOT_DIR/internal/build"

mkdir -p "$FAISS_BUILD_DIR"

cmake -S "$FAISS_SOURCE_DIR" -B "$FAISS_BUILD_DIR" \
    -DFAISS_ENABLE_GPU=OFF \
    -DFAISS_ENABLE_PYTHON=OFF \
    -DBUILD_SHARED_LIBS=OFF \
    -DFAISS_ENABLE_C_API=ON \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON

make -C "$FAISS_BUILD_DIR" -j faiss

mkdir -p "$FAISS_LIB_DIR"

cp "$FAISS_BUILD_DIR/faiss/libfaiss.a" "$FAISS_STATIC_LIB"

echo "Faiss static library built successfully."
