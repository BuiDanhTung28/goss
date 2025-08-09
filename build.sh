#!/bin/bash
# build.sh - Cross-platform build script

set -e # Exit immediately if a command exits with a non-zero status.

# --- Platform-specific dependency checks ---
OS="$(uname)"
echo "==> Detected OS: $OS"

if [[ "$OS" == "Linux" ]]; then
  # Check for cmake, make, and g++ on Linux
  if ! (command -v cmake &> /dev/null && command -v make &> /dev/null && command -v g++ &> /dev/null); then
    echo "=================================================================="
    echo "Error: 'cmake', 'make', and 'g++' are required for building."
    echo "On Debian/Ubuntu, run: sudo apt-get install build-essential cmake"
    echo "On Fedora/CentOS, run: sudo yum groupinstall 'Development Tools' && sudo yum install cmake"
    echo "=================================================================="
    exit 1
  fi
elif [[ "$OS" == "Darwin" ]]; then # macOS
  # Check for Xcode Command Line Tools
  if ! xcode-select -p &>/dev/null; then
    echo "=================================================================="
    echo "Error: Xcode Command Line Tools are required."
    echo "Please run 'xcode-select --install' in your terminal and try again."
    echo "=================================================================="
    exit 1
  fi
  # Check for Homebrew
  if ! command -v brew &> /dev/null; then
    echo "=================================================================="
    echo "Error: Homebrew is required for managing dependencies."
    echo "Please install it from https://brew.sh/"
    echo "=================================================================="
    exit 1
  fi
  # Check for cmake and libomp via Homebrew
  if ! (brew ls --versions cmake &> /dev/null && brew ls --versions libomp &> /dev/null); then
    echo "=================================================================="
    echo "Error: 'cmake' and 'libomp' are required."
    echo "Please run 'brew install cmake libomp' and try again."
    echo "=================================================================="
    exit 1
  fi
  # Add the Homebrew libomp path to CMAKE_PREFIX_PATH to help cmake find it
  export CMAKE_PREFIX_PATH=$(brew --prefix libomp):$CMAKE_PREFIX_PATH
  echo "==> Added Homebrew's libomp to CMAKE_PREFIX_PATH"
else
  echo "Warning: Unsupported OS '$OS'. Build might fail. Windows requires manual setup of MSVC or MinGW toolchain."
  # For now, we will proceed, but it's likely to fail without a proper toolchain.
fi

# --- Common build logic ---
ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
FAISS_LIB_DIR="$ROOT_DIR/internal/lib"
FAISS_STATIC_LIB="$FAISS_LIB_DIR/libfaiss.a"

# Only build if the static library does not exist to save time
if [ -f "$FAISS_STATIC_LIB" ]; then
    echo "==> Faiss static library already exists. Skipping build."
    exit 0
fi

echo "==> Building Faiss static library..."
FAISS_SOURCE_DIR="$ROOT_DIR/faiss_source"
FAISS_BUILD_DIR="$ROOT_DIR/internal/build"
mkdir -p "$FAISS_BUILD_DIR"

# Configure with CMake
cmake -S "$FAISS_SOURCE_DIR" -B "$FAISS_BUILD_DIR" \
    -DFAISS_ENABLE_GPU=OFF \
    -DFAISS_ENABLE_PYTHON=OFF \
    -DBUILD_SHARED_LIBS=OFF \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON

# Compile with Make
make -C "$FAISS_BUILD_DIR" -j faiss

# Create lib directory if it doesn't exist
mkdir -p "$FAISS_LIB_DIR"

# Copy the static library to the final location
cp "$FAISS_BUILD_DIR/faiss/libfaiss.a" "$FAISS_STATIC_LIB"
echo "==> Faiss static library built successfully." 