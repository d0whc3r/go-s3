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

// Package gos3 provides high-level S3 backup, listing, and retention cleanup
// operations on top of AWS SDK for Go v2.
package gos3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const maxDeleteBatchSize = 1000

// S3API defines the subset of AWS SDK v2 S3 client methods used by Manager.
// Accepting an interface enables clean unit testing with in-memory mocks.
type S3API interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

// UploadOptions configures behavior when uploading files or database dumps.
type UploadOptions struct {
	// Create creates the destination bucket if it does not already exist.
	Create bool
	// Replace overwrites an existing object at the destination key.
	Replace bool
	// ExpireDate sets an optional HTTP Expires header on uploaded objects.
	ExpireDate *time.Time
	// Compress zips the input files into a single archive before uploading.
	Compress bool
	// ZipName overrides the generated zip filename (implies Compress=true).
	ZipName string
}

// Manager performs S3 bucket and object operations.
type Manager struct {
	client S3API
	logger *slog.Logger
	now    func() time.Time
}

// New creates a new Manager backed by the given S3API client.
// If logger is nil, debug/info logs are discarded.
func New(client S3API, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Manager{
		client: client,
		logger: logger,
		now:    time.Now,
	}
}

// GetBuckets returns all buckets accessible by the configured credentials.
func (m *Manager) GetBuckets(ctx context.Context) ([]types.Bucket, error) {
	m.logger.Debug("listing buckets")
	out, err := m.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	return out.Buckets, nil
}

// CreateBucket creates an S3 bucket with the given name.
func (m *Manager) CreateBucket(ctx context.Context, bucket string) (*s3.CreateBucketOutput, error) {
	m.logger.Debug("creating bucket", "bucket", bucket)
	out, err := m.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("create bucket %q: %w", bucket, err)
	}
	return out, nil
}

// BucketExists reports whether the specified bucket exists and is accessible.
func (m *Manager) BucketExists(ctx context.Context, bucket string) bool {
	_, err := m.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	exists := err == nil
	m.logger.Debug("checked bucket existence", "bucket", bucket, "exists", exists)
	return exists
}

// RemoveBucket deletes a bucket. When force is true, all objects in the bucket
// are deleted first if the bucket is non-empty.
func (m *Manager) RemoveBucket(ctx context.Context, bucket string, force bool) (*s3.DeleteBucketOutput, error) {
	m.logger.Debug("removing bucket", "bucket", bucket, "force", force)
	out, err := m.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if force {
			if _, delErr := m.DeleteAllContent(ctx, bucket); delErr != nil {
				return nil, fmt.Errorf("empty bucket %q before deletion: %w", bucket, delErr)
			}
			return m.RemoveBucket(ctx, bucket, false)
		}
		return nil, fmt.Errorf("delete bucket %q: %w", bucket, err)
	}
	return out, nil
}

// ListFiles lists all objects in bucket matching an optional prefix, handling
// S3 pagination automatically.
func (m *Manager) ListFiles(ctx context.Context, bucket, prefix string) ([]types.Object, error) {
	m.logger.Debug("listing objects", "bucket", bucket, "prefix", prefix)
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	var objects []types.Object
	paginator := s3.NewListObjectsV2Paginator(m.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list objects in bucket %q: %w", bucket, err)
		}
		objects = append(objects, page.Contents...)
	}

	m.logger.Debug("listed objects", "bucket", bucket, "prefix", prefix, "count", len(objects))
	return objects, nil
}

// FileExists reports whether an object with the given key exists in bucket.
func (m *Manager) FileExists(ctx context.Context, bucket, key string) bool {
	_, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	exists := err == nil
	m.logger.Debug("checked object existence", "bucket", bucket, "key", key, "exists", exists)
	return exists
}

// DeleteAllContent deletes every object in bucket.
func (m *Manager) DeleteAllContent(ctx context.Context, bucket string) ([]types.DeletedObject, error) {
	files, err := m.ListFiles(ctx, bucket, "")
	if err != nil {
		return nil, err
	}
	return m.deleteObjects(ctx, bucket, files)
}

// UploadFile uploads a single local file to bucket under folder.
func (m *Manager) UploadFile(ctx context.Context, bucket, filePath, folder string, opts UploadOptions) (*s3.PutObjectOutput, error) {
	osFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", filePath, err)
	}
	defer osFile.Close()

	mimeType, err := detectContentType(osFile)
	if err != nil {
		return nil, fmt.Errorf("detect content type of %q: %w", filePath, err)
	}

	key := buildObjectKey(folder, filepath.Base(filePath))

	if opts.Create && !m.BucketExists(ctx, bucket) {
		if _, err := m.CreateBucket(ctx, bucket); err != nil {
			return nil, err
		}
	}

	if !opts.Replace && m.FileExists(ctx, bucket, key) {
		return nil, fmt.Errorf("file %q already exists in bucket %q and will not be replaced (use --replace to overwrite)", key, bucket)
	}

	m.logger.Debug("uploading file",
		"file", filePath,
		"bucket", bucket,
		"key", key,
		"content_type", mimeType,
	)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        osFile,
		Expires:     opts.ExpireDate,
		ContentType: aws.String(mimeType),
	}

	out, err := m.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("upload %q to s3://%s/%s: %w", filePath, bucket, key, err)
	}
	return out, nil
}

// UploadFiles uploads one or more local files, directories, or glob patterns to
// bucket under folder. When opts.Compress or opts.ZipName is set, all matched
// inputs are zipped into a single archive before uploading.
func (m *Manager) UploadFiles(ctx context.Context, bucket string, paths []string, folder string, opts UploadOptions) error {
	if len(paths) == 0 {
		return errors.New("no files specified for upload")
	}

	zipName := strings.TrimSpace(opts.ZipName)
	if zipName == "" && opts.Compress {
		zipName = fmt.Sprintf("zipped_%s.zip", m.now().Format("2006-01-02.150405"))
	}

	if zipName != "" {
		if !strings.HasSuffix(strings.ToLower(zipName), ".zip") {
			zipName += ".zip"
		}
		tmpDir, err := os.MkdirTemp("", "gos3-zip-*")
		if err != nil {
			return fmt.Errorf("create temp directory for zip: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		zipPath := filepath.Join(tmpDir, filepath.Base(zipName))
		m.logger.Debug("creating zip archive", "dest", zipPath, "sources", paths)
		if err := CreateZip(paths, zipPath); err != nil {
			return fmt.Errorf("compress files: %w", err)
		}

		_, err = m.UploadFile(ctx, bucket, zipPath, folder, opts)
		return err
	}

	files, err := collectFiles(paths)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("no files matched the provided paths")
	}

	for _, f := range files {
		if _, err := m.UploadFile(ctx, bucket, f, folder, opts); err != nil {
			return err
		}
	}
	return nil
}

// CleanOlder deletes objects in bucket (optionally restricted to folder) whose
// LastModified timestamp is older than durationExpr.
func (m *Manager) CleanOlder(ctx context.Context, bucket, durationExpr, folder string) ([]types.DeletedObject, error) {
	cutoff, err := ParseCutoffTime(m.now(), durationExpr)
	if err != nil {
		return nil, fmt.Errorf("parse duration %q: %w", durationExpr, err)
	}

	prefix := normalizeFolderPrefix(folder)
	m.logger.Debug("cleaning older objects",
		"bucket", bucket,
		"folder", folder,
		"prefix", prefix,
		"duration", durationExpr,
		"cutoff", cutoff.Format(time.RFC3339),
	)

	files, err := m.ListFiles(ctx, bucket, prefix)
	if err != nil {
		return nil, err
	}

	var toDelete []types.Object
	for _, f := range files {
		if f.LastModified != nil && f.LastModified.Before(cutoff) {
			toDelete = append(toDelete, f)
		}
	}

	m.logger.Debug("matched objects for deletion", "bucket", bucket, "matched", len(toDelete), "scanned", len(files))
	return m.deleteObjects(ctx, bucket, toDelete)
}

func (m *Manager) deleteObjects(ctx context.Context, bucket string, files []types.Object) ([]types.DeletedObject, error) {
	if len(files) == 0 {
		return nil, nil
	}

	var allDeleted []types.DeletedObject
	for start := 0; start < len(files); start += maxDeleteBatchSize {
		end := min(start+maxDeleteBatchSize, len(files))
		batch := files[start:end]

		identifiers := make([]types.ObjectIdentifier, 0, len(batch))
		for _, f := range batch {
			if f.Key == nil {
				continue
			}
			identifiers = append(identifiers, types.ObjectIdentifier{Key: f.Key})
		}
		if len(identifiers) == 0 {
			continue
		}

		m.logger.Debug("deleting object batch", "bucket", bucket, "count", len(identifiers))
		out, err := m.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{
				Objects: identifiers,
				Quiet:   aws.Bool(false),
			},
		})
		if err != nil {
			return allDeleted, fmt.Errorf("delete objects in bucket %q: %w", bucket, err)
		}
		allDeleted = append(allDeleted, out.Deleted...)
	}

	return allDeleted, nil
}

func buildObjectKey(folder, filename string) string {
	cleanFolder := strings.Trim(strings.ReplaceAll(folder, "\\", "/"), "/")
	if cleanFolder == "" {
		return filename
	}
	return path.Join(cleanFolder, filename)
}

func normalizeFolderPrefix(folder string) string {
	clean := strings.Trim(strings.ReplaceAll(folder, "\\", "/"), "/")
	if clean == "" {
		return ""
	}
	return clean + "/"
}

func detectContentType(f *os.File) (string, error) {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

func collectFiles(paths []string) ([]string, error) {
	expanded, err := expandSources(paths)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, p := range expanded {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %q: %w", p, err)
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(entryPath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !d.IsDir() {
				files = append(files, entryPath)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk directory %q: %w", p, err)
		}
	}
	return files, nil
}
