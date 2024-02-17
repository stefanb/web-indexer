# web-indexer

Quick and simple program to generate basic directory index pages from a local
file directory or S3 bucket, such as a file listing page.
Nothing revolutionary, but I needed something like this for a project or two.
It's generic enough to possibly be useful for others. S3 doesn't provide
file list indexes natively.

<p align="center">
  <img width="420" height="236" alt="screenshot" src=".github/readme/screenshot.png" />
</p>

This isn't a good solution for a dynamic listing of an S3 bucket (maybe try
Lambda), but it's simple and works for content that remains static between
deployments. If you're using Nginx, [fancyindex](https://www.nginx.com/resources/wiki/modules/fancy_index/)
does this dynamically with full customization. If you're using Apache,
[mod_autoindex](https://httpd.apache.org/docs/2.4/mod/mod_autoindex.html) is
what you're looking for, maybe with something like [Apaxy](https://oupala.github.io/apaxy/).

## Usage

Download from the [releases](https://github.com/joshbeard/web-indexer/releases)
or use the [`joshbeard/web-indexer`](https://hub.docker.com/r/joshbeard/web-indexer)
Docker image.

```plain
Usage:
  web-indexer --source <source> --target <target> [flags]

Flags:
  -u, --base-url string      A URL to prepend to the links
  -c, --config string        config file
      --date-format string   The date format to use in the index page (default "2006-01-02 15:04:05 MST")
  -h, --help                 help for web-indexer
  -i, --index-file string    The name of the index file (default "index.html")
  -l, --link-to-index        Link to the index file or just the path
  -F, --log-file string      The log file
  -L, --log-level string     The log level (default "info")
  -m, --minify               Minify the index page
  -q, --quiet                Suppress log output
  -r, --recursive            List files recursively
  -S, --skip strings         A list of files or directories to skip. Comma separated or specified multiple times
  -s, --source string        REQUIRED. The source directory or S3 URI to list
  -t, --target string        REQUIRED. The target directory or S3 URI to write to
  -f, --template string      A custom template file to use for the index page
  -T, --title string         The title of the index page
```

### Examples

The `source` and `target` arguments can be specified using their respective
flags, or provided as two unflagged arguments:

```shell
web-indexer SOURCE TARGET
web-indexer --source SOURCE --target TARGET
```

Index a local directory and write the index file to the same directory:

```shell
web-indexer --source /path/to/directory --target /path/to/directory
```

Index a local directory and write the index file to a different directory:

```shell
web-indexer --source /path/to/directory --target /foo/bar
```

Index a local directory and upload the index file to an S3 bucket:

```shell
web-indexer --source /path/to/directory --target s3://bucket/path
```

Index an S3 bucket and write the index file to a local directory:

```shell
web-indexer --source s3://bucket/path --target /path/to/directory
```

Index an S3 bucket and upload the index file to the same bucket and path:

```shell
web-indexer --source s3://bucket/path --target s3://bucket/path
```

Set a title for the index pages:

```shell
web-indexer --source /path/to/directory --target /path/to/directory --title 'Index of {relativePath}'
```

Load a config:

```shell
web-indexer -config /path/to/config.yaml
```

## GitHub Action

It's also available as a GitHub action.

For example:

```yaml
    - name: S3 Index Generator
      uses: joshbeard/web-indexer@0.1.1
      with:
        config: .web-indexer.yml
      env:
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        AWS_REGION: 'us-east-1'
```

Refer to the [`action.yml`](action.yml) for all available inputs, which
correspond to the CLI arguments and configuration parameters.

## Configuration

You can configure the behavior of `web-indexer` using command-line arguments
and/or a YAML config file. Both are evaluated with the command-line arguments
taking precedence.

```yaml
# base_url is an optional URL to prefix to links. If unset, links are relative.
base_url: ""

# date_format is the date format to use for indexed files modified time.
date_format: 2006-01-02 15:04:05 UTC

# index_file is the name of the file to generate.
index_file: index.html

# link_to_index toggles linking to the index_file for sub-paths or just the
# root of the subpath (foo/ vs foo/index.html).
link_to_index: false

# log_file is an optional path to a file to log to
# if log_level is set and this is not, messages are logged to stderr
log_file: ""

# log_level configures the verbosity of logging.
# Acceptable values: info, error, warn, debug
# Set this to an empty string or use the 'quiet' option  to suppress all log
# output.
log_level: "info"

# minify toggles minifying the generated HTML.
minify: false

# quiet suppresses all log output
quiet: false

# recursive enables indexing the source recursively.
recursive: false

# skips is a list of filenames to skip.
skips: []

# source is the path to a local directory or an S3 URI.
source: blah/

# target is the path to a local directory or an S3 URI.
target: blah/

# template is the path to a local Go template file to use for generating the
# indexes.
template: ""

# title customizes the title field available in the template.
# Certain tokens can be used to be dynamically replaced.
#   {source}       - the base source path
#   {target}       - the target path or URI
#   {path}         - the full path including the source
#   {relativePath} - the path relative to the source
title: ""
```

### Example Configuration

```yaml
source: /mnt/repos
target: s3://my-cool-repos/
title: "Index of {relativePath}"
minify: true
recursive: true
link_to_index: true
```
