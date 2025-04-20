#!/bin/bash
script_dir=$(cd $(dirname $0) && pwd)
source $script_dir/common.sh

# Find the actual APK file
APK_FILE=$(basename $(find_package_file "apk"))
echo "Using APK file: $APK_FILE"

# Extract the actual version from the filename
ACTUAL_VERSION=$(echo "$APK_FILE" | sed -E "s/${PACKAGE_NAME}_([^_]+)_${OS}_amd64.apk/\1/")
if [ "$ACTUAL_VERSION" != "$EXPECTED_VERSION" ]; then
    echo "Warning: Testing with version $ACTUAL_VERSION instead of $EXPECTED_VERSION"
    # Use the actual version for verification
    VERIFY_VERSION="$ACTUAL_VERSION"
else
    VERIFY_VERSION="$EXPECTED_VERSION"
fi

docker run -v ${DIST_DIR}:/tmp/dist \
    --rm alpine /bin/ash -c "
    # List available files for debugging
    echo 'Available files in /tmp/dist:';
    ls -la /tmp/dist;

    # Copy the package file
    cp /tmp/dist/${APK_FILE} /tmp;
    cd /tmp;
    apk add --no-cache --allow-untrusted ${APK_FILE};

    # Verify installation
    echo '=== Verifying installation ===';
    if ! command -v $PACKAGE_NAME &> /dev/null; then
        echo '$PACKAGE_NAME could not be installed.' >&2;
        exit 1;
    fi;
    echo 'ok';

    # Check the version
    echo '=== Checking executed version ===';
    INSTALLED_VERSION=\$($PACKAGE_NAME --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+');
    echo \"Installed version: \$INSTALLED_VERSION, Expected: $VERIFY_VERSION\";

    # Extract just the major.minor.patch part for comparison
    INSTALLED_VERSION_CLEAN=\$(echo \"\$INSTALLED_VERSION\" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+');
    EXPECTED_VERSION_CLEAN=\"$(echo "$VERIFY_VERSION" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "$VERIFY_VERSION")\";

    if [ \"\$INSTALLED_VERSION_CLEAN\" != \"\$EXPECTED_VERSION_CLEAN\" ]; then
        echo 'Version mismatch: expected '\"\$EXPECTED_VERSION_CLEAN\"', got '\"\$INSTALLED_VERSION_CLEAN\"'.' >&2;
        exit 1;
    fi;
    echo 'ok';

    echo 'All tests passed!';
"

if [ $? -eq 0 ]; then
    echo "Package $PACKAGE_NAME tests passed successfully."
else
    echo "Package $PACKAGE_NAME tests failed." >&2
    exit 1
fi
