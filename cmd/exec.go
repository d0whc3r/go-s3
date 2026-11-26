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

package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/gos3"
	"github.com/d0whc3r/go-s3/internal/version"
)

const tag = "[go-s3]"

func hasAction(o CLIOptions) bool {
	return o.Version || o.List || len(o.Backup) > 0 || len(o.Delete) > 0 || o.MySQL
}

func runCLI(cmd *cobra.Command, o CLIOptions, factory ClientFactory) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	if !hasAction(o) {
		return cmd.Help()
	}

	if o.Version {
		fmt.Fprintf(out, "v%s\n", version.Version)
		return nil
	}

	cfgOverrides := &config.Options{
		Bucket:   &o.Bucket,
		Endpoint: &o.Endpoint,
		Region:   &o.Region,
		Debug:    &o.Debug,
	}

	debugEnabled := config.Resolve(cfgOverrides).Debug
	var logger *slog.Logger
	if debugEnabled {
		logger = slog.New(slog.NewTextHandler(errOut, &slog.HandlerOptions{Level: slog.LevelDebug}))
		logger.Debug("debug logging enabled", "version", version.Version)
	} else {
		logger = slog.New(slog.DiscardHandler)
	}

	ctx := cmd.Context()
	client, cfg, err := factory(ctx, cfgOverrides, logger)
	if err != nil {
		return err
	}

	mgr := gos3.New(client, logger)
	uploadOpts := gos3.UploadOptions{
		Create:   o.Create,
		Replace:  o.Replace,
		Compress: o.Zip,
		ZipName:  o.ZipName,
	}

	if len(o.Delete) > 0 {
		if err := runDelete(cmd, mgr, cfg.Bucket, o.Delete, o.Folder, out); err != nil {
			return err
		}
	}

	if len(o.Backup) > 0 {
		if err := mgr.UploadFiles(ctx, cfg.Bucket, o.Backup, o.Folder, uploadOpts); err != nil {
			return err
		}
		fmt.Fprintf(out, "%s Backup success in bucket '%s'\n", tag, cfg.Bucket)
	}

	if o.MySQL {
		if err := mgr.UploadMySQL(ctx, cfg.Bucket, o.Folder, cfg.MySQL, uploadOpts); err != nil {
			return err
		}
		fmt.Fprintf(out, "%s MySQL dump backup success in bucket '%s'\n", tag, cfg.Bucket)
	}

	if o.List {
		if err := runList(cmd, mgr, cfg.Bucket, o.Folder, out); err != nil {
			return err
		}
	}

	return nil
}

func runDelete(cmd *cobra.Command, mgr *gos3.Manager, bucket string, deletes []string, defaultFolder string, out io.Writer) error {
	ctx := cmd.Context()
	for _, spec := range deletes {
		folder := defaultFolder
		durationExpr := spec
		if targetFolder, targetDuration, ok := strings.Cut(spec, "="); ok {
			folder = targetFolder
			durationExpr = targetDuration
		}

		deleted, err := mgr.CleanOlder(ctx, bucket, durationExpr, folder)
		if err != nil {
			return err
		}

		if folder != "" {
			fmt.Fprintf(out, "%s Deleted %d file(s) in folder '%s' older than '%s' in bucket '%s'\n", tag, len(deleted), folder, durationExpr, bucket)
		} else {
			fmt.Fprintf(out, "%s Deleted %d file(s) older than '%s' in bucket '%s'\n", tag, len(deleted), durationExpr, bucket)
		}
	}
	return nil
}

func runList(cmd *cobra.Command, mgr *gos3.Manager, bucket, folder string, out io.Writer) error {
	prefix := ""
	if trimmed := strings.Trim(folder, "/"); trimmed != "" {
		prefix = trimmed + "/"
	}
	files, err := mgr.ListFiles(cmd.Context(), bucket, prefix)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Fprintf(out, "%s No files found in bucket '%s'\n", tag, bucket)
		return nil
	}
	fmt.Fprintf(out, "%s File list in bucket '%s': %d\n", tag, bucket, len(files))
	for _, f := range files {
		if f.Key != nil {
			fmt.Fprintf(out, "%s\n", *f.Key)
		}
	}
	return nil
}
