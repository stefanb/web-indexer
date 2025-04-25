#!/bin/bash
script_dir=$(cd $(dirname $0) && pwd)

# This file should only be included, not executed.
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    echo "This script should be included, not executed."
    exit 1
fi

export PACKAGE_NAME="web-indexer"

# Allow overriding the version via environment variable
if [ -n "$TARGET_VERSION" ]; then
    export EXPECTED_VERSION="$TARGET_VERSION"
    echo "Using specified target version: $EXPECTED_VERSION"
else
    # Get the version with potential -dirty suffix
    export SOURCE_VERSION=$(git describe --tags --always --dirty)
    # Extract the clean version number
    export EXPECTED_VERSION=$(echo "$SOURCE_VERSION" | sed -E 's/([0-9]+\.[0-9]+\.[0-9]+).*/\1/')
    echo "Using git-derived version: $EXPECTED_VERSION (from $SOURCE_VERSION)"
fi

export OS="linux"
# Use the clean version for filename to match what GoReleaser produces
export FILENAME_BASE="${PACKAGE_NAME}_${EXPECTED_VERSION}_${OS}_amd64"
export DIST_DIR="${script_dir}/../../dist"
export TERM=xterm-256color

banner() {
    local text="$1"
    local width=${#text}
    local padding=$(( (60 - width) / 2 ))
    local stars=$(printf '%*s' "$padding" | tr ' ' '=')

    echo -e "\033[1;36m" # Cyan bold text
    echo "$stars $text $stars"
    echo -e "\033[0m" # Reset color
}

# Function to find the actual package file if the exact name doesn't match
find_package_file() {
    local extension=$1
    local exact_match="${DIST_DIR}/${FILENAME_BASE}.${extension}"

    # First try the exact match
    if [ -f "$exact_match" ]; then
        echo "$exact_match"
        return 0
    fi

    # If exact match not found, try to find a file with similar name
    # First try with the exact version
    local similar_file=$(find ${DIST_DIR} -name "${PACKAGE_NAME}_${EXPECTED_VERSION}_${OS}_amd64.${extension}" | head -n 1)

    if [ -n "$similar_file" ]; then
        echo "$similar_file"
        return 0
    fi

    # If still not found, try with any version
    similar_file=$(find ${DIST_DIR} -name "${PACKAGE_NAME}_*_${OS}_amd64.${extension}" | head -n 1)

    if [ -n "$similar_file" ]; then
        # Extract the actual version from the filename for verification
        local actual_version=$(basename "$similar_file" | sed -E "s/${PACKAGE_NAME}_([^_]+)_${OS}_amd64.${extension}/\1/")
        echo "Found package with version $actual_version instead of $EXPECTED_VERSION" >&2
        echo "$similar_file"
        return 0
    fi

    # Return the original name if nothing found (will fail gracefully later)
    echo "${FILENAME_BASE}.${extension}"
    return 1
}

export -f find_package_file
export -f banner
