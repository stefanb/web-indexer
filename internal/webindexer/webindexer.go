// Simple program to generate index files for a directory or S3 bucket.
package webindexer

import (
	_ "embed"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charmbracelet/log"
)

//go:embed templates/index.html.tmpl
var defaultTemplate string

// Indexer is the main struct for the webindexer package.
type Indexer struct {
	Cfg    Config
	Source FileSource
	Target FileSource
	s3     *s3.S3
}

// FileSource is an interface for listing the contents of a directory or S3
// bucket.
type FileSource interface {
	Read(path string) ([]Item, error)
	Write(data Data, content string) error
}

// Item represents an S3 key, or a local file/directory.
type Item struct {
	Name         string
	Size         string
	LastModified string
	URL          string
	IsDir        bool
	Items        []Item
}

// Data holds the template data.
type Data struct {
	Title        string
	Path         string
	RootPath     string
	RelativePath string
	URL          string
	Items        []Item
	Parent       string
	HasParent    bool
}

// New creates a new Indexer, taking the initial configuration and returning a
// updating it with the service, source and target paths.
func New(cfg Config) (*Indexer, error) {
	indexer := &Indexer{Cfg: cfg}

	if err := indexer.Cfg.Validate(); err != nil {
		return nil, err
	}

	if err := setupBackends(indexer); err != nil {
		return nil, err
	}

	return indexer, nil
}

func resolveParentPath(baseURL, parent, indexFile string, linkToIndexes bool) string {
	parentURL := strings.TrimSuffix(filepath.Join(baseURL, parent), "/") + "/"
	if linkToIndexes {
		parentURL += indexFile
	}
	return parentURL
}

func resolveItemURL(baseURL, path, name string, isDir, linkToIndexes bool, indexFile string) string {
	url := name
	if baseURL != "" {
		url = filepath.Join(baseURL, path, name)
	}

	if isDir {
		url = strings.TrimSuffix(url, "/") + "/"
		if linkToIndexes {
			url += indexFile
		}
	}
	return url
}

func sortItems(items *[]Item) {
	sort.SliceStable(*items, func(i, j int) bool {
		if (*items)[i].IsDir && !(*items)[j].IsDir {
			return true
		}
		if !(*items)[i].IsDir && (*items)[j].IsDir {
			return false
		}
		return (*items)[i].Name < (*items)[j].Name
	})
}

// setupBackends sets up the source and target backends for the indexer.
func setupBackends(indexer *Indexer) error {
	var err error

	if isS3URI(indexer.Cfg.Source) || isS3URI(indexer.Cfg.Target) {
		log.Debug("Setting up S3 session")
		sess, err := session.NewSession()
		if err != nil {
			return fmt.Errorf("failed to create AWS session: %w", err)
		}

		indexer.s3 = s3.New(sess)
	}

	indexer.Cfg.BasePath = strings.TrimSuffix(indexer.Cfg.Source, "/")
	if isS3URI(indexer.Cfg.Source) {
		_, prefix := uriToBucketAndPrefix(indexer.Cfg.Source)
		if prefix == "" {
			indexer.Cfg.BasePath = "/"
		} else {
			indexer.Cfg.BasePath = prefix
		}
	}

	indexer.Source, err = setupBackend(indexer.Cfg.Source, indexer)
	if err != nil {
		return err
	}

	indexer.Target, err = setupBackend(indexer.Cfg.Target, indexer)
	if err != nil {
		return err
	}

	return nil
}

// setupBackend sets up the backend for the given URI.
func setupBackend(uri string, indexer *Indexer) (FileSource, error) {
	log.Debugf("Setting up backend for %s", uri)
	if isS3URI(uri) {
		bucket, _ := uriToBucketAndPrefix(uri)
		return &S3Backend{svc: indexer.s3, bucket: bucket, cfg: indexer.Cfg}, nil
	}
	return &LocalBackend{path: uri, cfg: indexer.Cfg}, nil
}

// Generate the index file for the given path.
func (i Indexer) Generate(path string) error {
	var err error

	items, err := i.Source.Read(path)
	if err != nil {
		return err
	}

	data, err := i.data(items, path)
	if err != nil {
		return err
	}

	var tmpl *template.Template
	var templStr string
	if i.Cfg.Template != "" {
		log.Debugf("Using custom template %s for %s", i.Cfg.Template, path)
		templBytes, err := os.ReadFile(i.Cfg.Template)
		if err != nil {
			return err
		}
		templStr = string(templBytes)
	} else {
		log.Debugf("Using default template for %s", path)
		templStr = defaultTemplate
	}

	tmpl, err = template.New("index").Parse(templStr)
	if err != nil {
		return err
	}

	generated := new(strings.Builder)
	if err := tmpl.Execute(generated, data); err != nil {
		return err
	}

	output := generated.String()
	if i.Cfg.Minify {
		output = minifyHTML(generated.String())
	}

	if err := i.Target.Write(data, output); err != nil {
		return err
	}

	return nil
}

func (i Indexer) data(items []Item, path string) (Data, error) {
	relativePath := strings.TrimPrefix(path, i.Cfg.BasePath)

	// Ensure relative path is prefixed with a slash. This will also set an
	// empty base path to "/" (such as when listing the root of an S3 bucket).
	// S3 keys don't have a leading slash, but we normalize for consistency
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	data := Data{
		RootPath:     i.Cfg.BasePath,
		Items:        make([]Item, 0, len(items)),
		Path:         path,
		RelativePath: relativePath,
		URL:          i.Cfg.BaseURL,
		Title:        i.formatTitle(path, relativePath),
	}

	if path == i.Cfg.BasePath {
		data.HasParent = false
	} else {
		parent := filepath.Dir(path)
		data.Parent = resolveParentPath(i.Cfg.BaseURL, parent, i.Cfg.IndexFile, i.Cfg.LinkToIndexes)
		data.HasParent = parent != path
	}

	for _, item := range items {
		item, err := i.parseItem(path, item)
		if err != nil {
			return Data{}, err
		}

		data.Items = append(data.Items, item)
	}

	sortItems(&data.Items)
	return data, nil
}

func (i Indexer) parseItem(path string, item Item) (Item, error) {
	item.URL = resolveItemURL(i.Cfg.BaseURL, path, item.Name, item.IsDir, i.Cfg.LinkToIndexes, i.Cfg.IndexFile)

	if item.IsDir && i.Cfg.Recursive {
		if err := i.Generate(filepath.Join(path, item.Name)); err != nil {
			return Item{}, err
		}
	}
	return item, nil
}

func (i Indexer) formatTitle(path, relativePath string) string {
	title := strings.Replace(i.Cfg.Title, "{source}", filepath.Base(i.Cfg.Source), -1)
	title = strings.Replace(title, "{target}", filepath.Base(i.Cfg.Target), -1)
	title = strings.Replace(title, "{path}", path, -1)
	title = strings.Replace(title, "{relativePath}", relativePath, -1)

	return title
}

func humanizeBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	if bytes < 10 {
		return fmt.Sprintf("%d B", bytes)
	}
	log := math.Log(float64(bytes)) / math.Log(1024)
	index := int(log)
	size := float64(bytes) / math.Pow(1024, float64(index))
	return fmt.Sprintf("%.2f %s", size, units[index])
}

func minifyHTML(str string) string {
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "  ", "")
	str = strings.ReplaceAll(str, "> <", "><")
	return str
}

func shouldSkip(name, index string, skips []string) bool {
	if strings.HasSuffix(name, index) {
		return true
	}

	if contains(skips, name) {
		return true
	}

	return false
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
