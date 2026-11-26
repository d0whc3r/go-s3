// Copyright © 2020-2026 d0whc3r
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cmd defines the gos3 command-line interface.
package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/gos3"
	"github.com/d0whc3r/go-s3/internal/version"
)

// ClientFactory constructs an S3API client and resolved Configuration.
// Tests can override this to inject a mock S3 client.
type ClientFactory func(ctx context.Context, opts *config.Options, logger *slog.Logger) (gos3.S3API, config.Configuration, error)

// CLIOptions holds the parsed command-line flags.
type CLIOptions struct {
	Endpoint string
	Bucket   string
	Region   string
	List     bool
	Backup   []string
	Zip      bool
	ZipName  string
	Replace  bool
	Create   bool
	Folder   string
	Delete   []string
	MySQL    bool
	Debug    bool
	Version  bool
}

func defaultClientFactory(ctx context.Context, opts *config.Options, logger *slog.Logger) (gos3.S3API, config.Configuration, error) {
	return config.NewS3Client(ctx, opts, logger)
}

// NewRootCmd creates a new configured gos3 root Cobra command.
func NewRootCmd(factory ClientFactory) *cobra.Command {
	if factory == nil {
		factory = defaultClientFactory
	}

	var opts CLIOptions

	rootCmd := &cobra.Command{
		Use:           "gos3",
		Short:         fmt.Sprintf("S3-compatible backup and retention CLI (v%s)", version.Version),
		Long:          fmt.Sprintf("gos3 v%s — S3-compatible CLI utility for file backups, MySQL dumps, and retention cleanup.", version.Version),
		SilenceUsage:  true,
		SilenceErrors: false,
		Example: `  1. List files in "sample" bucket:
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -l

  2. Backup multiple files to "backupFolder":
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -b src/index.ts -b images/logo.png -f backupFolder

  3. Backup files using wildcard to "backup":
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -b "src/*" -b "images/*" -f backup

  4. Backup and zip into "zipped" (creating bucket if needed):
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -b "src/*" -z -f zipped -c

  5. Backup and zip with custom name, replacing existing archive:
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -b "src/*" -n allfiles.zip -f zipped -c -r

  6. Delete files older than 2 days in "uploads" and 1 month in "monthly":
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -d uploads=2d -d monthly=1M

  7. Delete files in "uploads" older than 30 minutes with debug output:
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -f uploads -d 30m --debug

  8. Dump MySQL database, zip it, and upload to "mysql-backup":
     $ gos3 -e https://s3.eu-central-1.amazonaws.com --bucket sample -f mysql-backup -m -z`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCLI(cmd, opts, factory)
		},
	}

	flags := rootCmd.Flags()
	flags.StringVarP(&opts.Endpoint, "endpoint", "e", "", "S3-compatible endpoint URL (or $ENDPOINT)")
	flags.StringVar(&opts.Bucket, "bucket", "", "Destination bucket name (or $BUCKET)")
	flags.StringVar(&opts.Region, "region", "", "AWS/S3 region (or $REGION, default: us-east-1)")
	flags.BoolVarP(&opts.List, "list", "l", false, "List all files in the bucket")
	flags.StringArrayVarP(&opts.Backup, "backup", "b", nil, "File, directory, or glob pattern to backup (repeatable)")
	flags.BoolVarP(&opts.Zip, "zip", "z", false, "Compress backup files into a zip archive")
	flags.StringVarP(&opts.ZipName, "zipname", "n", "", "Custom zip filename (implies --zip)")
	flags.BoolVarP(&opts.Replace, "replace", "r", false, "Overwrite existing files in S3")
	flags.BoolVarP(&opts.Create, "create", "c", false, "Create destination bucket if it does not exist")
	flags.StringVarP(&opts.Folder, "folder", "f", "", "Destination folder prefix in the bucket")
	flags.StringArrayVarP(&opts.Delete, "delete", "d", nil, "Delete files older than duration (e.g. 2d, 1M, or folder=7d; repeatable)")
	flags.BoolVarP(&opts.MySQL, "mysql", "m", false, "Create and upload a MySQL dump ($MYSQL_USER, $MYSQL_PASSWORD, $MYSQL_DATABASE, $MYSQL_HOST, $MYSQL_PORT)")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable verbose debug logging to stderr (or $DEBUG=true)")
	flags.BoolVarP(&opts.Version, "version", "v", false, "Print gos3 version and exit")

	return rootCmd
}

// Execute runs the default root command with the provided context.
func Execute(ctx context.Context) error {
	return NewRootCmd(nil).ExecuteContext(ctx)
}
