package webindexer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charmbracelet/log"
)

type S3Backend struct {
	svc    *s3.S3
	bucket string
	cfg    Config
}

func (s *S3Backend) Read(prefix string) ([]Item, error) {
	// Ensure the prefix has a trailing slash for s3 keys
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Remove leading slash for s3 keys
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix == "/" {
		prefix = ""
	}

	log.Debugf("Listing objects in %s/%s", s.bucket, prefix)

	req := &s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	resp, err := s.svc.ListObjectsV2(req)
	if err != nil {
		return nil, fmt.Errorf("unable to list S3 objects: %w", err)
	}

	var items []Item
	for _, content := range resp.Contents {
		if shouldSkip(*content.Key, s.cfg.IndexFile, s.cfg.Skips) {
			continue
		}

		itemName := filepath.Base(*content.Key)
		item := Item{
			Name:         itemName,
			Size:         humanizeBytes(*content.Size),
			LastModified: content.LastModified.Format(s.cfg.DateFormat),
			IsDir:        false,
		}

		items = append(items, item)
	}

	for _, commonPrefix := range resp.CommonPrefixes {
		dirName := strings.TrimPrefix(*commonPrefix.Prefix, prefix)
		item := Item{
			Name:  dirName,
			IsDir: true,
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *S3Backend) Write(data Data, content string) error {
	bucket, target := uriToBucketAndPrefix(s.cfg.Target)
	target = strings.TrimPrefix(target, s.cfg.BasePath)
	target = filepath.Join(target, data.RelativePath, s.cfg.IndexFile)

	strReader := strings.NewReader(content)
	size := humanizeBytes(int64(strReader.Len()))
	log.Infof("Uploading %s to %s/%s", size, bucket, target)

	_, err := s.svc.PutObject(&s3.PutObjectInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(target),
		Body:            aws.ReadSeekCloser(strReader),
		ContentType:     aws.String("text/html"),
		ContentEncoding: aws.String("utf-8"),
	})
	return err
}

func isS3URI(uri string) bool {
	return strings.HasPrefix(uri, "s3://")
}

func uriToBucketAndPrefix(uri string) (string, string) {
	uri = strings.TrimPrefix(uri, "s3://")
	uriParts := strings.SplitN(uri, "/", 2)

	if len(uriParts) == 1 {
		return uriParts[0], ""
	}

	return uriParts[0], uriParts[1]
}
