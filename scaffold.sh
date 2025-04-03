#!/bin/bash

# Create test directory structure
echo "Creating test directory structure..."

# Create base test directory
TEST_DIR="./web-indexer-test"
rm -rf $TEST_DIR
mkdir -p $TEST_DIR

# Regular directories with content
mkdir -p $TEST_DIR/regular/subdir1
mkdir -p $TEST_DIR/regular/subdir2/subsubdir
echo "This is a sample file" > $TEST_DIR/regular/file1.txt
echo "Another sample file" > $TEST_DIR/regular/subdir1/file2.txt
echo "Deep nested file" > $TEST_DIR/regular/subdir2/subsubdir/file3.txt

# Directory with .noindex file - should be completely skipped
mkdir -p $TEST_DIR/noindex/subdir1
echo "This is hidden content" > $TEST_DIR/noindex/secret.txt
echo "More hidden content" > $TEST_DIR/noindex/subdir1/secret2.txt
touch $TEST_DIR/noindex/.noindex

# Directory with .skipindex file - should appear in parent but not be indexed itself
mkdir -p $TEST_DIR/skipindex/subdir1
echo "This is a custom website" > $TEST_DIR/skipindex/index.html
echo "Custom website subpage" > $TEST_DIR/skipindex/subdir1/page.html
touch $TEST_DIR/skipindex/.skipindex

# Directory with existing index.html treated as skipindex file
mkdir -p $TEST_DIR/website/css
mkdir -p $TEST_DIR/website/js
mkdir -p $TEST_DIR/website/images
echo '<!DOCTYPE html><html><head><title>My Website</title></head><body><h1>Welcome to my website</h1></body></html>' > $TEST_DIR/website/index.html
echo 'body { font-family: sans-serif; }' > $TEST_DIR/website/css/style.css
echo 'console.log("Hello world");' > $TEST_DIR/website/js/main.js

# Make target directory
mkdir -p $TEST_DIR-output

echo "Test directories created at $TEST_DIR"
echo "Output directory created at $TEST_DIR-output"

echo ""
echo "Run web-indexer with:"
echo "./web-indexer --source $TEST_DIR --target $TEST_DIR-output --recursive --noindex-files .noindex --skipindex-files .skipindex,index.html"
echo ""
echo "To test skipindex-files with only .skipindex (not index.html):"
echo "./web-indexer --source $TEST_DIR --target $TEST_DIR-output --recursive --noindex-files .noindex --skipindex-files .skipindex"
echo ""
echo "Directory structure:"
find $TEST_DIR -type f | sort
