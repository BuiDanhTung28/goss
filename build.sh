#!/bin/bash
# build.sh - Cross-platform build script for goss

set -e

# --- MAIN DIRECTORIES ---
ROOT_DIR=$(cd "$(dirname "$0")" && pwd)
FAISS_SOURCE_DIR="$ROOT_DIR/faiss_source"

# --- OS DETECTION ---
OS="$(uname)"
echo "==> Detected OS: $OS"

if [[ "$OS" == "Linux" ]]; then
  echo "==> Building for Linux using Docker..."

  # Build the Docker image. This Dockerfile should build a complete libfaiss.a
  docker build -t faiss-builder .

  # Create a temporary container and copy the library file
  CONTAINER_ID=$(docker create faiss-builder)
  mkdir -p "$ROOT_DIR/internal/lib/linux_amd64"
  docker cp "$CONTAINER_ID":/libfaiss.a "$ROOT_DIR/internal/lib/linux_amd64/libfaiss.a"
  docker rm "$CONTAINER_ID"

elif [[ "$OS" == "Darwin" ]]; then # macOS
  echo "==> Building for macOS locally..."

  # --- CHECK MACOS DEPENDENCIES ---
  if ! (command -v cmake &> /dev/null && command -v g++ &> /dev/null && command -v brew &> /dev/null && brew ls --versions libomp &> /dev/null); then
    echo "=================================================================="
    echo "Error: Required dependencies (cmake, g++, brew, libomp) not found."
    echo "Please run 'brew install cmake libomp' and ensure Xcode Command Line Tools are installed."
    echo "=================================================================="
    exit 1
  fi
  
  # Add libomp path to CMAKE_PREFIX_PATH for cmake
  export CMAKE_PREFIX_PATH=$(brew --prefix libomp):$CMAKE_PREFIX_PATH
  echo "==> Added Homebrew's libomp to CMAKE_PREFIX_PATH"

  # --- MACOS-SPECIFIC BUILD PROCESS ---
  FAISS_BUILD_DIR="$ROOT_DIR/internal/build/darwin_arm64"
  FAISS_LIB_PATH="$ROOT_DIR/internal/lib/darwin_arm64/libfaiss.a"

  # Build only if the library file does not exist
  if [ -f "$FAISS_LIB_PATH" ]; then
      echo "==> Faiss static library for macOS already exists. Skipping build."
      exit 0
  fi

  echo "==> Building Faiss static library for macOS..."
  mkdir -p "$FAISS_BUILD_DIR"
  
  # Configure with CMake
  cmake -S "$FAISS_SOURCE_DIR" -B "$FAISS_BUILD_DIR" \
      -DFAISS_ENABLE_GPU=OFF \
      -DFAISS_ENABLE_PYTHON=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_BUILD_TYPE=Release \
      -DCMAKE_POSITION_INDEPENDENT_CODE=ON \
      -DFAISS_BUILD_TESTING=OFF

  # Chỉ biên dịch target "faiss" để tạo thư viện chính
  cmake --build "$FAISS_BUILD_DIR" --target faiss --config Release
  
  # Create the destination directory and copy the file
  mkdir -p "$ROOT_DIR/internal/lib/darwin_arm64"
  cp "$FAISS_BUILD_DIR/faiss/libfaiss.a" "$FAISS_LIB_PATH"
  echo "==> Faiss static library for macOS built successfully."

else
  echo "Warning: Unsupported OS '$OS'. Build might fail. Windows requires manual setup of MSVC or MinGW toolchain."
  exit 1
fi

echo "==> Build process completed successfully!"