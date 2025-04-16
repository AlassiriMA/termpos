#!/bin/bash

# Script to build TermPOS for multiple platforms
# This script will create binaries for Linux, Windows, and macOS

set -e # Exit immediately if a command exits with a non-zero status

# Define variables
BINARY_NAME="termpos"
VERSION=$(grep -o 'Version     = "[^"]*"' cmd/pos/version.go | cut -d'"' -f2)
BUILD_DATE=$(date +%Y-%m-%d)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}"
OUTPUT_DIR="./bin"

# Create output directory if it doesn't exist
mkdir -p "${OUTPUT_DIR}"
echo "Building TermPOS v${VERSION} (${BUILD_DATE})"
echo "----------------------------------------"

# Function to build for a specific platform
build() {
    local GOOS=$1
    local GOARCH=$2
    local SUFFIX=$3
    local TARGET_DIR="${OUTPUT_DIR}/${GOOS}_${GOARCH}"
    local BINARY="${TARGET_DIR}/${BINARY_NAME}${SUFFIX}"
    
    mkdir -p "${TARGET_DIR}"
    
    echo "Building for ${GOOS}/${GOARCH}..."
    GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "${LDFLAGS}" -o "${BINARY}" ./cmd/pos
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Success: ${BINARY}"
    else
        echo "  ✗ Failed: ${BINARY}"
        return 1
    fi
    
    return 0
}

# Build for different platforms
echo "Building for Linux (amd64)..."
build "linux" "amd64" ""

echo "Building for Windows (amd64)..."
build "windows" "amd64" ".exe"

echo "Building for macOS (amd64)..."
build "darwin" "amd64" ""

# Build for ARM platforms (optional)
if [ "$1" == "--with-arm" ]; then
    echo "Building for Linux (arm64)..."
    build "linux" "arm64" ""
    
    echo "Building for macOS (arm64)..."
    build "darwin" "arm64" ""
fi

echo "----------------------------------------"
echo "Build completed successfully!"
echo "Binaries available in ${OUTPUT_DIR}/"