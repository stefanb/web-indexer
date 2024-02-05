// Simple program to generate index.html files for an S3 bucket.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"
)

//go:embed index.html.tmpl
var defaultTemplate string

type Config struct {
	Bucket        string `yaml:"bucket"`
	Prefix        string `yaml:"prefix"`
	Title         string `yaml:"title"`
	Upload        bool   `yaml:"upload"`
	URL           string `yaml:"url"`
	LinkToIndexes bool   `yaml:"link_to_index"`
	RelativeLinks bool   `yaml:"relative_links"`
	StagingDir    string `yaml:"staging_dir"`
	DateFormat    string `yaml:"date_format"`
	Debug         bool   `yaml:"debug"`
	Template      string `yaml:"template"`
	cfgFile       string `yaml:"-"`
}

var cfg Config

// Item represents a file in the S3 bucket.
type Item struct {
	Name         string
	Size         string
	LastModified string
	URL          string
	IsDir        bool
}

// Data holds the template data.
type Data struct {
	Title     string
	Items     []Item
	Parent    string
	HasParent bool
}

func main() {
	flag.StringVar(&cfg.cfgFile, "config", "", "The path to an optional config file")
	flag.StringVar(&cfg.Bucket, "bucket", "", "The name of the S3 bucket")
	flag.StringVar(&cfg.Prefix, "prefix", "", "The path within the bucket to list")
	flag.StringVar(&cfg.Title, "title", "", "The title of the index page")
	flag.BoolVar(&cfg.Upload, "upload", false, "Upload a file to the S3 bucket")
	flag.StringVar(&cfg.URL, "url", "", "The URL of the S3 bucket")
	flag.BoolVar(&cfg.LinkToIndexes, "link-to-index", false,
		"Link to index.html or just the path")
	flag.BoolVar(&cfg.RelativeLinks, "relative-links", false,
		"Use relative links instead of absolute links")
	flag.StringVar(&cfg.StagingDir, "staging-dir", "_staging",
		"The directory to use for staging files (when not uploading)")
	flag.StringVar(&cfg.DateFormat, "date-format", "2006-01-02 15:04:05 MST",
		"The date format to use in the index page")
	flag.BoolVar(&cfg.Debug, "debug", false, "Print debug information")
	flag.StringVar(&cfg.Template, "template", "",
		"A custom template file to use for the index page")
	flag.Parse()

	// Set the configuration based on the command-line flags and the optional
	// config file. The CLI flags take precedence over the config file.
	setConfig()

	debug("Config: %+v\n", cfg)

	if cfg.Bucket == "" || cfg.URL == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg.Prefix = strings.Trim(cfg.Prefix, "/")
	if cfg.Prefix != "" && !strings.HasSuffix(cfg.Prefix, "/") {
		cfg.Prefix += "/"
	}

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	// Generate index.html files for each directory
	generateIndexes(svc, cfg.Bucket, cfg.Prefix, cfg.URL)

	fmt.Println("Index pages generated successfully.")
}

func generateIndexes(svc *s3.S3, bucket, prefix, url string) {
	req := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	resp, err := svc.ListObjectsV2(req)
	if err != nil {
		panic(err)
	}

	// Prepare data for the template
	var items []Item
	for _, content := range resp.Contents {
		if *content.Key == prefix || strings.HasSuffix(*content.Key, "index.html") {
			continue
		}

		itemName := filepath.Base(*content.Key)
		item := Item{
			Name:         itemName,
			Size:         humanizeBytes(*content.Size),
			LastModified: content.LastModified.Format(cfg.DateFormat),
			IsDir:        false,
		}

		if cfg.RelativeLinks {
			item.URL = itemName
		} else {
			item.URL = url + "/" + *content.Key
		}

		items = append(items, item)
	}

	for _, commonPrefix := range resp.CommonPrefixes {
		dirName := strings.TrimPrefix(*commonPrefix.Prefix, prefix)
		item := Item{
			Name:  dirName,
			IsDir: true,
		}

		if cfg.RelativeLinks {
			item.URL = dirName
		} else {
			item.URL = url + "/" + *commonPrefix.Prefix
		}

		if cfg.LinkToIndexes {
			item.URL += "index.html"
		}

		items = append(items, item)
	}

	data := Data{
		Title: cfg.Title + " - " + prefix,
		Items: items,
	}

	parent := filepath.Dir(strings.TrimSuffix(prefix, "/"))
	if prefix != "" {
		data.HasParent = true
	}

	if cfg.RelativeLinks {
		data.Parent = "../"
	} else {
		data.Parent = url + "/" + parent
	}

	if cfg.LinkToIndexes {
		data.Parent += "index.html"
	}

	// Sort directories first, then files by their last modified date.
	// This is a simple way to make the index page more readable.
	sort.SliceStable(data.Items, func(i, j int) bool {
		if data.Items[i].IsDir && !data.Items[j].IsDir {
			return true
		}
		if !data.Items[i].IsDir && data.Items[j].IsDir {

			return false
		}
		return data.Items[i].Name < data.Items[j].Name
	})

	var tmpl *template.Template
	if cfg.Template != "" {
		debug("Using custom template %s\n", cfg.Template)
		tmplStr, err := os.ReadFile(cfg.Template)
		if err != nil {
			panic(err)
		}
		tmpl, err = template.New("index").Parse(string(tmplStr))
		if err != nil {
			panic(err)
		}
	} else {
		debug("Using default template\n")
		tmpl, err = template.New("index").Parse(defaultTemplate)
		if err != nil {
			panic(err)
		}
	}

	// Generate and optionally upload each index.html
	for _, cp := range resp.CommonPrefixes {
		generateIndexes(svc, bucket, *cp.Prefix, url)
	}

	generateOrUploadIndex(svc, bucket, prefix, data, tmpl)
}

func generateOrUploadIndex(svc *s3.S3, bucket, prefix string, data Data, tmpl *template.Template) {
	if cfg.Upload {
		wr := new(strings.Builder)
		if err := tmpl.Execute(wr, data); err != nil {
			panic(err)
		}

		wrStr := minifyHTML(wr.String())

		debug("Uploading index.html to %s/%sindex.html\n", bucket, prefix)
		if err := uploadToS3(svc, bucket, wrStr, prefix+"index.html"); err != nil {
			panic(err)
		}
	} else {
		// Define the local directory structure based on the prefix
		localPath := filepath.Join(cfg.StagingDir, bucket, prefix)
		if err := os.MkdirAll(localPath, 0755); err != nil {
			panic(fmt.Errorf("failed to create directory %s: %w", localPath, err))
		}

		filePath := filepath.Join(localPath, "index.html")
		file, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			panic(err)
		}

		fmt.Printf("Index page generated successfully at %s\n", filePath)
	}
}

func uploadToS3(svc *s3.S3, bucket, content, dst string) error {
	r := strings.NewReader(content)
	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(dst),
		Body:            aws.ReadSeekCloser(r),
		ContentType:     aws.String("text/html"),
		ContentEncoding: aws.String("utf-8"),
	})
	return err
}

// HumanizeBytes converts bytes to a human-readable format.
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

// setConfig reads an optional setConfig file and overwrites the default values with
// command-line flags.
// This sets the global args variable.
func setConfig() {
	config := Config{}
	if cfg.cfgFile != "" {
		config = readConfig()
	}

	// Overwrite config file values with command-line flags
	if cfg.Bucket != "" {
		config.Bucket = cfg.Bucket
	}

	if cfg.Prefix != "" {
		config.Prefix = cfg.Prefix
	}

	if cfg.Title != "" {
		config.Title = cfg.Title
	}

	if cfg.URL != "" {
		config.URL = cfg.URL
	}

	if cfg.LinkToIndexes {
		config.LinkToIndexes = cfg.LinkToIndexes
	}

	if cfg.RelativeLinks {
		config.RelativeLinks = cfg.RelativeLinks
	}

	if cfg.StagingDir != "" {
		config.StagingDir = cfg.StagingDir
	}

	if cfg.DateFormat != "" {
		config.DateFormat = cfg.DateFormat
	}

	if cfg.Debug {
		config.Debug = cfg.Debug
	}

	if cfg.Upload {
		config.Upload = cfg.Upload
	}

	if cfg.Template != "" {
		config.Template = cfg.Template
	}

	cfg = config
}

func readConfig() Config {
	config := Config{}
	// Check if the config file exists
	if _, err := os.Stat(cfg.cfgFile); os.IsNotExist(err) {
		fmt.Printf("Config file %s does not exist\n", cfg.cfgFile)
		os.Exit(1)
	}

	// Read the config file.
	fmt.Printf("Reading configuration from %s\n", cfg.cfgFile)
	file, err := os.Open(cfg.cfgFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return config
}

func minifyHTML(str string) string {
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\t", "")
	str = strings.ReplaceAll(str, "  ", "")
	str = strings.ReplaceAll(str, "> <", "><")
	return str
}

func debug(msg string, args ...interface{}) {
	if cfg.Debug {
		fmt.Printf(msg, args...)
	}
}
