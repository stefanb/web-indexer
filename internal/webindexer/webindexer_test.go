package webindexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBackendSetup struct {
	mock.Mock
}

type MockSource struct {
	mock.Mock
}

func (m *MockSource) Read(path string) ([]Item, bool, error) {
	args := m.Called(path)
	return args.Get(0).([]Item), args.Bool(1), args.Error(2)
}

func (m *MockSource) Write(data Data, content string) error {
	args := m.Called(data, content)
	return args.Error(0)
}

func TestIndexer_Generate(t *testing.T) {
	mockSource := new(MockSource)
	mockTarget := new(MockSource)
	indexer := Indexer{
		Source: mockSource,
		Target: mockTarget,
		Cfg: Config{
			Minify: true,
		},
	}

	mockSource.On("Read", mock.Anything).Return([]Item{}, false, nil)
	mockTarget.On("Write", mock.Anything, mock.Anything).Return(nil)

	err := indexer.Generate("path/to/generate")
	assert.NoError(t, err)

	mockSource.AssertExpectations(t)
	mockTarget.AssertExpectations(t)
}

func TestCustomTemplate(t *testing.T) {
	mockSource := new(MockSource)
	mockTarget := new(MockSource)

	// Create a temporary file for the custom template
	file, err := os.CreateTemp("", "template.html")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	// Write the custom template to the file
	html := "<html><head><title>Custom Template</title></head></html>"
	_, err = file.WriteString(html)
	require.NoError(t, err)

	// Close the file
	err = file.Close()
	require.NoError(t, err)

	// Create an indexer with the custom template
	indexer := Indexer{
		Source: mockSource,
		Target: mockTarget,
		Cfg: Config{
			Template: file.Name(),
			Minify:   true,
		},
	}

	require.NoError(t, err)

	mockSource.On("Read", mock.Anything).Return([]Item{}, false, nil)
	mockTarget.On("Write", mock.Anything, mock.Anything).Return(nil)

	err = indexer.Generate("path/to/generate")
	assert.NoError(t, err)

	// Assert that the custom template was used and the content is the same
	args := mockTarget.Calls[0].Arguments
	assert.Equal(t, html, args.Get(1))

	// Check the file content
	f, err := os.ReadFile(file.Name())
	require.NoError(t, err)
	assert.Equal(t, html, string(f))

	mockSource.AssertExpectations(t)
	mockTarget.AssertExpectations(t)
}

func (m *MockBackendSetup) Setup(indexer *Indexer) error {
	args := m.Called(indexer)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	mockSetup := new(MockBackendSetup)
	cfg := Config{
		Source: "s3://mybucket/some/source/path",
		Target: "/mnt/some/target/path",
		SortBy: "name",
		Order:  "asc",
	}

	mockSetup.On("Setup", mock.Anything).Return(nil)

	indexer, err := New(cfg)
	require.NoError(t, err)
	indexer.BackendSetup = mockSetup

	err = indexer.BackendSetup.Setup(indexer)
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Source: "some/source/path",
				Target: "some/target/path",
				SortBy: "name",
				Order:  "asc",
				Minify: true,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing source",
			config: Config{
				Target: "some/target/path",
				SortBy: "name",
				Order:  "asc",
			},
			wantErr: true,
		},
		{
			name: "invalid config - missing target",
			config: Config{
				Source: "some/source/path",
				SortBy: "name",
				Order:  "asc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMinifyHTML(t *testing.T) {
	html := `
<html>
  <head>
    <title>Test</title>
  </head>
</html>
`
	minified := minifyHTML(html)

	// The minified HTML should not contain any newlines
	assert.False(t, strings.Contains(minified, "\n"))

	// The minified HTML should not contain any leading or trailing whitespace
	assert.False(t, strings.HasPrefix(minified, " "))
	assert.False(t, strings.HasSuffix(minified, " "))

	// The minified HTML should not contain any whitespace between tags
	assert.False(t, strings.Contains(minified, "> <"))
}

func TestShouldSkipURL(t *testing.T) {
	// URLs that should be skipped
	skips := []string{
		"index.html",
		"index.htm",
		"foo.txt",
	}

	assert.True(t, shouldSkip("index.html", "index.html", skips))
	assert.True(t, shouldSkip("index.htm", "index.htm", skips))
	assert.True(t, shouldSkip("foo.txt", "index.html", skips))
	assert.False(t, shouldSkip("another-file.tar.gz", "index.html", skips))
	assert.False(t, shouldSkip("something.html", "index.html", skips))
}

func TestResolveParentPath(t *testing.T) {
	tests := []struct {
		baseURL       string
		parent        string
		indexFile     string
		linkToIndexes bool
		expected      string
	}{
		{"", "dir", "index.html", false, "dir"},
		{"/blah", "dir", "index.html", false, "/blah/dir"},
		{"https://foo.com/one/two", "dir", "index.html", false, "https://foo.com/one/two/dir"},
		{"https://foo.com/one/two/", "dir", "index.html", false, "https://foo.com/one/two/dir"},
	}

	for _, test := range tests {
		result := resolveParentPath(test.baseURL, test.parent, test.indexFile, test.linkToIndexes)
		assert.Equal(t, test.expected, result)
	}
}

func TestResolveItemURL(t *testing.T) {
	tests := []struct {
		baseURL       string
		path          string
		name          string
		isDir         bool
		linkToIndexes bool
		indexFile     string
		expected      string
	}{
		{"https://foo.com", "one/two", "three", false, false, "", "https://foo.com/one/two/three"},
		{"https://foo.com", "one/two", "three/", true, false, "", "https://foo.com/one/two/three/"},
		{"https://foo.com", "one/two", "three", true, true, "index.html", "https://foo.com/one/two/three/index.html"},
		// Test without base URL
		{"", "one/two", "three", false, false, "", "three"},
		// Test with empty path
		{"https://foo.com", "", "three", false, false, "", "https://foo.com/three"},
		// Error scenario can be mocked if joinURL is mockable or by causing it to fail with specific input
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolveItemURL(tc.baseURL, tc.path, tc.name, tc.isDir, tc.linkToIndexes, tc.indexFile)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIndexer_ParseItem(t *testing.T) {
	// Create some temporary directories and files
	sourceDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(sourceDir) // Clean up after the test

	// Create mock files
	fileNames := []string{"file1.txt", "file2.txt"}
	for _, fName := range fileNames {
		tmpFn := filepath.Join(sourceDir, fName)
		if err := os.WriteFile(tmpFn, []byte("test content"), 0o666); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
	}
	// Create a mock directory
	dirName := "dir"
	if err := os.Mkdir(filepath.Join(sourceDir, dirName), 0o755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	targetDir, err := os.MkdirTemp("", "test")
	defer os.RemoveAll(targetDir)
	require.NoError(t, err)

	indexer := Indexer{
		Cfg: Config{
			BaseURL:       "https://example.com",
			Source:        sourceDir,
			Target:        targetDir,
			LinkToIndexes: true,
			IndexFile:     "index.html",
			Recursive:     true,
		},
	}
	err = setupBackends(&indexer)
	require.NoError(t, err)

	// Simulate items
	itemDir := Item{Name: filepath.Join(sourceDir, "dir"), IsDir: true}
	itemFile := Item{Name: "file.html", IsDir: false}

	// Test directory item
	modifiedItemDir, err := indexer.parseItem("", itemDir)
	require.NoError(t, err)
	expectUrl := "https://example.com/tmp/" + filepath.Base(sourceDir) + "/dir/index.html"
	assert.Equal(t, expectUrl, modifiedItemDir.URL)

	// Test file item
	modifiedItemFile, err := indexer.parseItem("dir", itemFile)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/dir/file.html", modifiedItemFile.URL)
}
