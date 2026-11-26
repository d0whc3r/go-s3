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

package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/d0whc3r/go-s3/cmd"
	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/gos3"
	"github.com/d0whc3r/go-s3/internal/version"
)

type cliMockObj struct {
	body         []byte
	lastModified time.Time
}

type cliMockS3 struct {
	buckets map[string]map[string]cliMockObj
}

func newCLIMockS3(bucket string) *cliMockS3 {
	return &cliMockS3{
		buckets: map[string]map[string]cliMockObj{
			bucket: {},
		},
	}
}

func (m *cliMockS3) ListBuckets(_ context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{}, nil
}

func (m *cliMockS3) CreateBucket(_ context.Context, p *s3.CreateBucketInput, _ ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	m.buckets[aws.ToString(p.Bucket)] = map[string]cliMockObj{}
	return &s3.CreateBucketOutput{}, nil
}

func (m *cliMockS3) DeleteBucket(_ context.Context, p *s3.DeleteBucketInput, _ ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	delete(m.buckets, aws.ToString(p.Bucket))
	return &s3.DeleteBucketOutput{}, nil
}

func (m *cliMockS3) HeadBucket(_ context.Context, p *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	if _, ok := m.buckets[aws.ToString(p.Bucket)]; !ok {
		return nil, errors.New("not found")
	}
	return &s3.HeadBucketOutput{}, nil
}

func (m *cliMockS3) ListObjectsV2(_ context.Context, p *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	objs, ok := m.buckets[aws.ToString(p.Bucket)]
	if !ok {
		return nil, errors.New("no such bucket")
	}
	prefix := aws.ToString(p.Prefix)
	var keys []string
	for k := range objs {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var contents []types.Object
	for _, k := range keys {
		contents = append(contents, types.Object{
			Key:          aws.String(k),
			LastModified: aws.Time(objs[k].lastModified),
		})
	}
	return &s3.ListObjectsV2Output{Contents: contents}, nil
}

func (m *cliMockS3) HeadObject(_ context.Context, p *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if _, ok := m.buckets[aws.ToString(p.Bucket)][aws.ToString(p.Key)]; !ok {
		return nil, errors.New("not found")
	}
	return &s3.HeadObjectOutput{}, nil
}

func (m *cliMockS3) PutObject(_ context.Context, p *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	data, err := io.ReadAll(p.Body)
	if err != nil {
		return nil, err
	}
	m.buckets[aws.ToString(p.Bucket)][aws.ToString(p.Key)] = cliMockObj{
		body:         data,
		lastModified: time.Now(),
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *cliMockS3) DeleteObjects(_ context.Context, p *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	objs := m.buckets[aws.ToString(p.Bucket)]
	var deleted []types.DeletedObject
	for _, id := range p.Delete.Objects {
		k := aws.ToString(id.Key)
		if _, ok := objs[k]; ok {
			delete(objs, k)
			deleted = append(deleted, types.DeletedObject{Key: aws.String(k)})
		}
	}
	return &s3.DeleteObjectsOutput{Deleted: deleted}, nil
}

func runTestCLI(t *testing.T, mock gos3.S3API, args ...string) (string, string, error) {
	t.Helper()
	factory := func(_ context.Context, opts *config.Options, _ *slog.Logger) (gos3.S3API, config.Configuration, error) {
		cfg := config.Resolve(opts)
		if cfg.Bucket == "" {
			cfg.Bucket = "test-bucket"
		}
		return mock, cfg, nil
	}
	root := cmd.NewRootCmd(factory)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func TestCLIHelpAndVersion(t *testing.T) {
	mock := newCLIMockS3("test-bucket")

	stdout, _, err := runTestCLI(t, mock)
	if err != nil {
		t.Fatalf("expected nil error for help, got: %v", err)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Errorf("expected usage output, got: %q", stdout)
	}

	stdout, _, err = runTestCLI(t, mock, "-v")
	if err != nil {
		t.Fatalf("unexpected error for -v: %v", err)
	}
	want := fmt.Sprintf("v%s\n", version.Version)
	if stdout != want {
		t.Errorf("version output = %q, want %q", stdout, want)
	}
}

func TestCLIBackupListDeleteAndDebug(t *testing.T) {
	mock := newCLIMockS3("test-bucket")
	tmp := t.TempDir()
	sample := filepath.Join(tmp, "note.txt")
	if err := os.WriteFile(sample, []byte("backup data"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Backup with --debug enabled.
	stdout, stderr, err := runTestCLI(t, mock, "--bucket", "test-bucket", "-b", sample, "-f", "docs", "--debug")
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if !strings.Contains(stdout, "[go-s3] Backup success in bucket 'test-bucket'") {
		t.Errorf("unexpected stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "level=DEBUG") {
		t.Errorf("expected debug logs on stderr, got: %q", stderr)
	}

	// List files.
	stdout, _, err = runTestCLI(t, mock, "--bucket", "test-bucket", "-l")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(stdout, "docs/note.txt") {
		t.Errorf("expected docs/note.txt in list output, got: %q", stdout)
	}

	// Make object appear 5 days old and delete with folder=2d.
	oldObj := mock.buckets["test-bucket"]["docs/note.txt"]
	oldObj.lastModified = time.Now().AddDate(0, 0, -5)
	mock.buckets["test-bucket"]["docs/note.txt"] = oldObj

	stdout, _, err = runTestCLI(t, mock, "--bucket", "test-bucket", "-d", "docs=2d")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !strings.Contains(stdout, "Deleted 1 file(s) in folder 'docs' older than '2d'") {
		t.Errorf("unexpected delete stdout: %q", stdout)
	}

	// List should now report no files.
	stdout, _, err = runTestCLI(t, mock, "--bucket", "test-bucket", "-l")
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if !strings.Contains(stdout, "No files found") {
		t.Errorf("expected empty bucket message, got: %q", stdout)
	}
}
