#!/bin/bash
script_dir=$(cd $(dirname $0) && pwd)

# Check if a version is specified
if [ "$1" == "--version" ] || [ "$1" == "-v" ]; then
    export TARGET_VERSION="$2"
    shift 2
fi

source $script_dir/common.sh

FAILED=0
FAILED_TESTS=""
PASSED_TESTS=""

green=`tput setaf 2`
red=`tput setaf 1`
reset=`tput sgr0`

tests=$(ls $script_dir/test-*.sh)

if [ -n "$1" ]; then
    tests=$(ls $script_dir/test-$1*.sh)
fi

if [ -z "$tests" ]; then
    echo "No tests found in $script_dir"
    exit 1
fi

for test_script in $tests; do
    banner "[$(basename $test_script)]"
    if ! bash $test_script; then
        echo "${red}❌ FAILED: $(basename "${test_script}")${reset}"
        FAILED=1
        FAILED_TESTS="$FAILED_TESTS $(basename "$test_script")"
        continue
    fi

    echo "${green}✅ PASSED: $(basename $test_script)${reset}"
    PASSED_TESTS="$PASSED_TESTS $(basename $test_script)"
done

echo
echo "==============================================================================="
if [ $FAILED -eq 0 ]; then
    echo "${green}✅ All tests passed!${reset}"
else
    if [ -n "$PASSED_TESTS" ]; then
        echo "${green}✅ Passing tests:$PASSED_TESTS${reset}"
    fi
    echo "${red}❌ Failing tests:$FAILED_TESTS${reset}"
    exit 1
fi
