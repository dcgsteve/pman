#!/bin/bash

# Build script for pman - creates binaries for multiple platforms and architectures

set -e

CLI_VERSION=$(grep 'Version = ' cli/main.go | sed 's/.*Version = "\(.*\)".*/\1/')
SERVER_VERSION=$(grep 'Version = ' backend/main.go | sed 's/.*Version = "\(.*\)".*/\1/')
BINARY_NAME="pman"
RELEASES_DIR="releases"

# Define build targets (OS:ARCH)
declare -a BUILD_TARGETS=(
    "linux:amd64"
    "linux:arm64"
    "windows:amd64"
    "darwin:amd64"
    "darwin:arm64"
)

# Define components to build
declare -a COMPONENTS=(
    "server:backend/main.go"
    "cli:cli/main.go"
)

echo "Building $BINARY_NAME - CLI version: $CLI_VERSION, Server version: $SERVER_VERSION"

# Create releases directory
mkdir -p "$RELEASES_DIR"

# Clean previous builds
rm -rf "$RELEASES_DIR"/*

build_binary() {
    local os=$1
    local arch=$2
    local component_name=$3
    local source_path=$4
    
    local output_name="${BINARY_NAME}-${component_name}-${os}-${arch}"
    
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    local output_path="$RELEASES_DIR/$output_name"
    
    echo "Building $component_name for $os/$arch..."
    
    # Determine version based on component
    local version=""
    if [ "$component_name" = "cli" ]; then
        version=$CLI_VERSION
    elif [ "$component_name" = "server" ]; then
        version=$SERVER_VERSION
    fi
    
    # Pure Go build for all components (now that we're using modernc.org/sqlite)
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.Version=$version -s -w" \
        -a -installsuffix nocgo \
        -o "$output_path" \
        "$source_path"
    
    if [ $? -eq 0 ]; then
        echo "✓ Successfully built: $output_name"
    else
        echo "✗ Failed to build: $output_name"
        return 1
    fi
}

# Build all combinations
for target in "${BUILD_TARGETS[@]}"; do
    IFS=':' read -r os arch <<< "$target"
    
    for component in "${COMPONENTS[@]}"; do
        IFS=':' read -r component_name source_path <<< "$component"
        build_binary "$os" "$arch" "$component_name" "$source_path"
    done
done

echo ""
echo "Build complete! Binaries available in $RELEASES_DIR:"
ls -la "$RELEASES_DIR"

echo ""
echo "Summary:"
echo "- CLI Version: $CLI_VERSION"
echo "- Server Version: $SERVER_VERSION"
echo "- Platforms: ${#BUILD_TARGETS[@]}"
echo "- Components: ${#COMPONENTS[@]}"
echo "- Total binaries: $((${#BUILD_TARGETS[@]} * ${#COMPONENTS[@]}))"
