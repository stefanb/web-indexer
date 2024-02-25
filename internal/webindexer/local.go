package webindexer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

type LocalBackend struct {
	path string
	cfg  Config
}

func (l *LocalBackend) Read(path string) ([]Item, error) {
	var items []Item
	log.Debugf("Listing files in %s", path)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read source path %s: %w", path, err)
	}

	for _, file := range files {
		if shouldSkip(file.Name(), l.cfg.IndexFile, l.cfg.Skips) {
			continue
		}

		stat, err := os.Stat(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("unable to stat file %s: %w", file.Name(), err)
		}

		size := humanizeBytes(stat.Size())
		modified := stat.ModTime().Format(l.cfg.DateFormat)

		itemName := file.Name()
		item := Item{
			Name:         itemName,
			Size:         size,
			LastModified: modified,
			IsDir:        file.IsDir(),
		}

		items = append(items, item)
	}

	return items, nil
}

func (l *LocalBackend) Write(data Data, content string) error {
	prefix := data.RelativePath
	prefix = strings.TrimPrefix(prefix, l.cfg.BasePath)
	localPath := filepath.Join(l.cfg.Target, prefix)
	if err := os.MkdirAll(localPath, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", localPath, err)
	}

	filePath := filepath.Join(localPath, l.cfg.IndexFile)
	file, err := os.Create(filePath) // #nosec
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	log.Infof("Generated %s", filePath)
	return nil
}
