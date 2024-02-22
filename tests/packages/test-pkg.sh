#!/bin/bash
script_dir=$(cd $(dirname $0) && pwd)
source $script_dir/common.sh

docker run -v ${DIST_DIR}:/tmp/dist \
    --rm rockylinux:9 /bin/bash -c "
    cp /tmp/dist/${FILENAME_BASE}.tar.gz /tmp;
    tar -xzf /tmp/${FILENAME_BASE}.tar.gz -C /tmp;
    mv /tmp/${PACKAGE_NAME} /usr/local/bin/${PACKAGE_NAME};
    chmod +x /usr/local/bin/${PACKAGE_NAME};

    # Verify installation
    echo '=== Verifying installation ===';
    if ! command -v $PACKAGE_NAME &> /dev/null; then
        echo '$PACKAGE_NAME could not be installed.' >&2;
        exit 1;
    fi;
    echo 'ok';

    # Check the version

    echo '=== Checking executed version ===';
    INSTALLED_VERSION=\$($PACKAGE_NAME --version | grep -oP '\d+\.\d+\.\d+');
    if [ \"\$INSTALLED_VERSION\" != \"$EXPECTED_VERSION\" ]; then
        echo 'Version mismatch: expected $EXPECTED_VERSION, got '\"\$INSTALLED_VERSION\"'.' >&2;
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
