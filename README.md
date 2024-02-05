# s3-index-generator

Quick and simple program to generate basic directory index pages for S3
buckets to provide a browseable file listing of paths. Nothing revolutionary,
but I needed something like this for a project or two. It's generic enough to
possibly be useful for others.

<p align="center">
  <img width="420" height="236" alt="screenshot" src="https://github.com/joshbeard/s3-index-generator/blob/main/.github/readme/screenshot.png" />
</p>

The static files generated should be uploaded to the S3 bucket, which can be
done using the `-upload` flag.

This isn't a good solution for a dynamic listing of a bucket (maybe try
Lambda), but it's simple and works for content that remains static between
deployments.

## Usage

Download from the [releases](https://github.com/joshbeard/s3-index-generator/releases)
or use the [`joshbeard/s3-index-generator`](https://hub.docker.com/r/joshbeard/s3-index-generator)
Docker image.

```plain
Usage of s3-index-generator:
  -bucket string
    	The name of the S3 bucket
  -config string
    	The path to an optional config file
  -date-format string
    	The date format to use in the index page (default "2006-01-02 15:04:05 MST")
  -debug
    	Print debug information
  -link-to-index
    	Link to index.html or just the path
  -prefix string
    	The path within the bucket to list
  -relative-links
    	Use relative links instead of absolute links
  -staging-dir string
    	The directory to use for staging files (when not uploading) (default "_staging")
  -template string
    	A custom template file to use for the index page
  -title string
    	The title of the index page
  -upload
    	Upload a file to the S3 bucket
  -url string
    	The URL of the S3 bucket
```

__Example:__

```shell
s3-index-generator -config .s3-index.yml -upload
```

## GitHub Action

It's also available as a GitHub action.

For example:

```yaml
    - name: S3 Index Generator
      uses: joshbeard/s3-index-generator@v1
      with:
        config: .s3-indexer.yml
        upload: "true"
      env:
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        AWS_REGION: 'us-east-1'
```

Refer to the [`action.yml`](action.yml) for all available inputs, which
correspond to the CLI arguments and configuration parameters.

## Configuration

You can configure the behavior of `s3-index-generator` using command-line arguments
and/or a YAML config file. Both are evaluated with the command-line arguments
taking precedence.

```yaml
# the name of the s3 bucket
bucket: my-bucket

# prefix is the path/key to start at in the bucket
prefix: "/"

# link to index will link to "index.html" for paths instead of just the path.
# e.g. foo/ vs foo/index.html
link_to_index: true

# toggles using relative links or absolute links
relative_links: true

# title is shown at the top of the pages
title: get.example.app

# url is used for generating links when relative_links is false
url: https://get.example.app

# staging_dir is used when running without the '-upload' flag.
staging_dir: _s3_index
```
