#!/bin/bash
script_dir=$(cd $(dirname $0) && pwd)

# This file should only be included, not executed.
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    echo "This script should be included, not executed."
    exit 1
fi

export PACKAGE_NAME="web-indexer"
export SOURCE_VERSION=$(git describe --tags --always --dirty)
export EXPECTED_VERSION=$(echo "$SOURCE_VERSION" | sed -E 's/([0-9]+\.[0-9]+\.[0-9]+).*/\1/')

export OS="linux"
export FILENAME_BASE="${PACKAGE_NAME}_${SOURCE_VERSION}_${OS}_amd64"
export DIST_DIR="${script_dir}/../../dist"
export TERM=xterm-256color

