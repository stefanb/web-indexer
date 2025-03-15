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
	items, hasNoIndex, err := localBackend.Read(tempDir)
	if err != nil {
		t.Errorf("Failed to read directory: %v", err)
	}
	assert.False(t, hasNoIndex)

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

func TestLocalBackendReadWithNoIndex(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files including a noindex file
	files := map[string]string{
		"file1.txt": "test content",
		".noindex":  "",
	}
	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Setup LocalBackend with noindex configuration
	localBackend := LocalBackend{
		path: tempDir,
		cfg: Config{
			DateFormat:   "2006-01-02",
			IndexFile:    "index.html",
			NoIndexFiles: []string{".noindex"},
		},
	}

	// Test reading directory with noindex file
	items, hasNoIndex, err := localBackend.Read(tempDir)
	require.NoError(t, err)
	assert.True(t, hasNoIndex)
	assert.Empty(t, items)
}

func TestLocalBackendReadSubdirWithNoIndex(t *testing.T) {
	// Setup temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a subdirectory with a noindex file
	subDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	// Create test files
	files := map[string]string{
		"file1.txt":        "test content",
		"subdir/file2.txt": "test content",
		"subdir/.noindex":  "",
	}
	for name, content := range files {
		err := os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0o644)
		require.NoError(t, err)
	}

	// Setup LocalBackend with noindex configuration
	localBackend := LocalBackend{
		path: tempDir,
		cfg: Config{
			DateFormat:   "2006-01-02",
			IndexFile:    "index.html",
			NoIndexFiles: []string{".noindex"},
		},
	}

	// Test reading parent directory
	items, hasNoIndex, err := localBackend.Read(tempDir)
	require.NoError(t, err)
	assert.False(t, hasNoIndex)
	assert.Len(t, items, 1) // Should only contain file1.txt, not subdir

	// Test reading subdirectory with noindex file
	items, hasNoIndex, err = localBackend.Read(subDir)
	require.NoError(t, err)
	assert.True(t, hasNoIndex)
	assert.Empty(t, items)
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
