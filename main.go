package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joshbeard/web-indexer/internal/webindexer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version is set at build time using -ldflags
var version = "dev"

func main() {
	var cfg webindexer.Config
	rootCmd := &cobra.Command{
		Use:     "web-indexer [flags]",
		Version: version,
		Short:   "Generate index files for a directory or S3 bucket",
		Long: "web-indexer is a tool to generate index files for a directory or S3 bucket.\n\n" +
			"See https://github.com/joshbeard/web-indexer for more information.\n\n" +
			"The source and target can be specified using their flags or as " +
			"the\nfirst and second arguments.\n\nA custom configuration file can be " +
			"specified using the --config flag.\nBy default, web-indexer will look " +
			"for a .web-indexer.yml or .web-indexer.yaml\nfile in the current directory.",
		Example: strings.Join([]string{
			"  Index a local directory and write the index file to the same directory",
			"    web-indexer --source /path/to/directory --target /path/to/directory",
			"  Index a local directory and write the index file to a different directory",
			"    web-indexer --source /path/to/directory --target /foo/bar",
			"  Index a local directory and upload the index file to an S3 bucket",
			"    web-indexer --source /path/to/directory --target s3://bucket/path",
			"  Index an S3 bucket and write the index file to a local directory",
			"    web-indexer --source s3://bucket/path --target /path/to/directory",
			"  Index an S3 bucket and upload the index file to the same bucket and path",
			"    web-indexer --source s3://bucket/path --target s3://bucket/path",
			"",
			"  Run with a custom configuration file",
			"    web-indexer -c custom.yml /path/to/source /path/to/target",
			"",
			"  Set a title for the index pages",
			"    web-indexer \\",
			"      --source /path/to/directory \\",
			"      --target /path/to/directory \\",
			"      --title 'Index of {relativePath}'",
		}, "\n"),
		Run: func(cmd *cobra.Command, args []string) {
			err := viper.Unmarshal(&cfg)
			cobra.CheckErr(err)

			// If 2 arguments are passed, the first is the source and the
			// second is the target
			if len(args) == 2 {
				cfg.Source = args[0]
				cfg.Target = args[1]
			} else if len(args) > 0 {
				log.Fatalf("Unknown arguments: %s", args)
			}

			if err := setupLogger(cfg); err != nil {
				log.Fatalf("Failed to setup logger: %s", err)
			}

			indexer, err := webindexer.New(cfg)
			cobra.CheckErr(err)

			log.Infof("Generating index for %s", cfg.Source)
			err = indexer.Generate(indexer.Cfg.BasePath)
			cobra.CheckErr(err)
		},
	}

	err := setFlags(rootCmd, &cfg)
	cobra.CheckErr(err)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}

func initConfig(cfgFile *string) func() {
	return func() {
		if *cfgFile != "" {
			viper.SetConfigFile(*cfgFile)
		} else {
			// Look for a .web-indexer.yml or .web-indexer.yaml file in the
			// current directory
			for _, name := range []string{".web-indexer.yml", ".web-indexer.yaml"} {
				if _, err := os.Stat(name); err == nil {
					viper.SetConfigFile(name)
					break
				}
			}
		}

		// Environment variables
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err == nil {
			log.Debugf("Using config file: %s", viper.ConfigFileUsed())
		}
	}
}

func setFlags(rootCmd *cobra.Command, cfg *webindexer.Config) error {
	cobra.OnInitialize(initConfig(&cfg.CfgFile))

	rootCmd.PersistentFlags().StringVarP(&cfg.CfgFile, "config", "c", "", "config file")

	rootCmd.Flags().StringVarP(&cfg.BaseURL, "base-url", "u", "", "A URL to prepend to the links")
	if err := viper.BindPFlag("base-url", rootCmd.Flags().Lookup("base-url")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.DateFormat, "date-format", "", "2006-01-02 15:04:05 MST", "The date format to use in the index page")
	if err := viper.BindPFlag("date-format", rootCmd.Flags().Lookup("date-format")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.IndexFile, "index-file", "i", "index.html", "The name of the index file")
	if err := viper.BindPFlag("index-file", rootCmd.Flags().Lookup("index-file")); err != nil {
		return err
	}

	rootCmd.Flags().BoolVarP(&cfg.LinkToIndexes, "link-to-index", "l", false, "Link to the index file or just the path")
	if err := viper.BindPFlag("link-to-index", rootCmd.Flags().Lookup("link-to-index")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.LogLevel, "log-level", "L", "info", "The log level")
	if err := viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.LogFile, "log-file", "F", "", "The log file")
	if err := viper.BindPFlag("log-file", rootCmd.Flags().Lookup("log-file")); err != nil {
		return err
	}

	rootCmd.Flags().BoolVarP(&cfg.Minify, "minify", "m", false, "Minify the index page")
	if err := viper.BindPFlag("minify", rootCmd.Flags().Lookup("minify")); err != nil {
		return err
	}

	rootCmd.Flags().BoolVarP(&cfg.Quiet, "quiet", "q", false, "Suppress log output")
	if err := viper.BindPFlag("quiet", rootCmd.Flags().Lookup("quiet")); err != nil {
		return err
	}

	rootCmd.Flags().BoolVarP(&cfg.Recursive, "recursive", "r", false, "List files recursively")
	if err := viper.BindPFlag("recursive", rootCmd.Flags().Lookup("recursive")); err != nil {
		return err
	}

	rootCmd.Flags().StringSliceVarP(&cfg.Skips, "skip", "S", []string{}, "A list of files or directories to skip. "+
		"Comma separated or specified multiple times")
	if err := viper.BindPFlag("skip", rootCmd.Flags().Lookup("skip")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.Source, "source", "s", "", "REQUIRED. The source directory or S3 URI to list")
	if err := viper.BindPFlag("source", rootCmd.Flags().Lookup("source")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.Target, "target", "t", "", "REQUIRED. The target directory or S3 URI to write to")
	if err := viper.BindPFlag("target", rootCmd.Flags().Lookup("target")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.Template, "template", "f", "", "A custom template file to use for the index page")
	if err := viper.BindPFlag("template", rootCmd.Flags().Lookup("template")); err != nil {
		return err
	}

	rootCmd.Flags().StringVarP(&cfg.Title, "title", "T", "", "The title of the index page")
	if err := viper.BindPFlag("title", rootCmd.Flags().Lookup("title")); err != nil {
		return err
	}

	return nil
}

func setupLogger(cfg webindexer.Config) error {
	if cfg.LogLevel == "" || cfg.Quiet {
		devnull, err := os.Open(os.DevNull)
		if err != nil {
			return err
		}

		log.SetOutput(devnull)

		return nil
	}

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("unable to parse log level: %w", err)
	}

	log.SetLevel(logLevel)

	if cfg.LogFile == "" {
		log.SetOutput(os.Stdout)

		return nil
	}

	f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	} else {
		log.SetOutput(f)
	}

	return nil
}
