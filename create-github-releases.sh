#!/bin/bash

# GitHub Release Creation Script for pman
# Creates separate releases for server and CLI binaries

set -e

# Get versions from source files
CLI_VERSION=$(grep 'Version = ' cli/main.go | sed 's/.*Version = "\(.*\)".*/\1/')
SERVER_VERSION=$(grep 'Version = ' backend/main.go | sed 's/.*Version = "\(.*\)".*/\1/')

echo "CLI Version: $CLI_VERSION"
echo "Server Version: $SERVER_VERSION"

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Please install it from: https://cli.github.com/"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    exit 1
fi

# Function to create a release
create_release() {
    local tag=$1
    local title=$2
    local notes=$3
    local files=$4
    
    echo ""
    echo "Creating release: $title"
    echo "Tag: $tag"
    
    # Fetch latest tags from remote
    echo "Fetching latest tags from remote..."
    git fetch --tags --prune --prune-tags
    
    # Check if tag already exists locally or remotely
    if git rev-parse "$tag" >/dev/null 2>&1; then
        echo "Error: Tag $tag already exists locally. Release for this version has already been created."
        exit 1
    fi
    
    # Check if tag exists on remote
    if git ls-remote --tags origin | grep -q "refs/tags/$tag$"; then
        echo "Error: Tag $tag already exists on remote. Release for this version has already been created."
        exit 1
    fi
    
    # Create and push tag
    git tag -a "$tag" -m "$title"
    git push origin "$tag"
    
    # Create GitHub release
    gh release create "$tag" \
        --title "$title" \
        --notes "$notes" \
        $files
    
    echo "✓ Release created successfully: $title"
}

# Build binaries if not already built
if [ ! -d "releases" ] || [ -z "$(ls -A releases 2>/dev/null)" ]; then
    echo "Building binaries..."
    ./build-binaries.sh
fi

# Prepare release notes
CLI_NOTES="## pman CLI v$CLI_VERSION

Password Manager CLI tool for secure password storage and retrieval.

### Installation

Download the appropriate binary for your platform and make it executable:

\`\`\`bash
chmod +x pman-cli-*
\`\`\`

### Usage

\`\`\`bash
./pman-cli login -s https://your-server.com -u username -p password
./pman-cli add myproject/api_key \"secret_value\"
./pman-cli get myproject/api_key
./pman-cli ls
\`\`\`

For more information, see the [documentation](https://github.com/dcgsteve/pman#readme)."

SERVER_NOTES="## pman Server v$SERVER_VERSION

Password Manager Server for secure password storage backend.

### Installation

Download the appropriate binary for your platform and make it executable:

\`\`\`bash
chmod +x pman-server-*
\`\`\`

### Usage

Set required environment variables:
\`\`\`bash
export PMAN_ENCRYPTION_KEY=\"your-32-character-encryption-key\"
export PMAN_DOMAIN_NAME=\"localhost:5000\"
export PMAN_DEFAULT_EXPIRE_DAYS=\"30\"
\`\`\`

Run the server:
\`\`\`bash
./pman-server-*
\`\`\`

For more information, see the [documentation](https://github.com/dcgsteve/pman#readme)."

# Create CLI release
CLI_FILES=""
for file in releases/pman-cli-*; do
    if [ -f "$file" ]; then
        CLI_FILES="$CLI_FILES $file"
    fi
done

if [ -n "$CLI_FILES" ]; then
    create_release "cli-v$CLI_VERSION" "pman CLI v$CLI_VERSION" "$CLI_NOTES" "$CLI_FILES"
else
    echo "Warning: No CLI binaries found in releases/"
fi

# Create Server release
SERVER_FILES=""
for file in releases/pman-server-*; do
    if [ -f "$file" ]; then
        SERVER_FILES="$SERVER_FILES $file"
    fi
done

if [ -n "$SERVER_FILES" ]; then
    create_release "server-v$SERVER_VERSION" "pman Server v$SERVER_VERSION" "$SERVER_NOTES" "$SERVER_FILES"
else
    echo "Warning: No server binaries found in releases/"
fi

echo ""
echo "Release creation complete!"
echo ""
echo "View releases at: https://github.com/dcgsteve/pman/releases"