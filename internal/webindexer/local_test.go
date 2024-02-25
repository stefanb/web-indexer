package webindexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalBackendRead(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create mock files
	fileNames := []string{"file1.txt", "file2.txt"}
	for _, fName := range fileNames {
		tmpFn := filepath.Join(tempDir, fName)
		if err := os.WriteFile(tmpFn, []byte("test content"), 0o644); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}

	// Setup LocalBackend and Config
	localBackend := LocalBackend{
		path: tempDir,
		cfg: Config{
			DateFormat: "2006-01-02",
			IndexFile:  "index.html",
			Skips:      []string{"skip"},
		},
	}

	// Test the Read function
	items, err := localBackend.Read(tempDir)
	if err != nil {
		t.Errorf("Failed to read directory: %v", err)
	}

	if len(items) != len(fileNames) {
		t.Errorf("Expected %d items, got %d", len(fileNames), len(items))
	}

	// Verify file items
	for _, item := range items {
		if !contains(fileNames, item.Name) {
			t.Errorf("Unexpected file: %s", item.Name)
		}
	}
}

func TestLocalBackendWrite(t *testing.T) {
	// Setup temporary directory for target
	targetDir, err := os.MkdirTemp("", "target")
	require.NoError(t, err, "Failed to create temp target dir")
	defer os.RemoveAll(targetDir) // Clean up after the test

	// Setup LocalBackend with test configuration
	localBackend := LocalBackend{
		cfg: Config{
			BasePath:  "/base",
			Target:    targetDir,
			IndexFile: "index.html",
		},
	}

	// Define test data
	data := Data{
		RelativePath: "/base/subdir/",
	}
	content := "<html>Test Content</html>"

	// Execute the Write method
	err = localBackend.Write(data, content)
	require.NoError(t, err, "Failed to write content")

	// Verify the file and its content
	filePath := filepath.Join(targetDir, "subdir", localBackend.cfg.IndexFile)
	_, err = os.Stat(filePath)
	require.NoError(t, err, "File was not created")

	readContent, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read file")

	assert.Equal(t, strings.TrimSpace(content), strings.TrimSpace(string(readContent)), "File content does not match")
}
