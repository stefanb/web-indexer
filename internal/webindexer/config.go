package webindexer

import (
	"fmt"
)

type Config struct {
	BaseURL       string   `yaml:"base_url"`
	DateFormat    string   `yaml:"date_format"`
	IndexFile     string   `yaml:"index_file"`
	LinkToIndexes bool     `yaml:"link_to_index"`
	LogLevel      string   `yaml:"log_level"`
	LogFile       string   `yaml:"log_file"`
	Minify        bool     `yaml:"minify"`
	Quiet         bool     `yaml:"quiet"`
	Recursive     bool     `yaml:"recursive"`
	Skips         []string `yaml:"skips"`
	Source        string   `yaml:"source"`
	Target        string   `yaml:"target"`
	Template      string   `yaml:"template"`
	Title         string   `yaml:"title"`
	CfgFile       string   `yaml:"-"`
	BasePath      string   `yaml:"-"`
}

func (c Config) Validate() error {
	if c.Source == "" {
		return fmt.Errorf("source is required")
	}

	if c.Target == "" {
		return fmt.Errorf("target is required")
	}

	return nil
}
