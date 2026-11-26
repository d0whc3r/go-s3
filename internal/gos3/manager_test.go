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

package gos3_test

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/gos3"
)

type mockObject struct {
	body         []byte
	contentType  string
	lastModified time.Time
}

type mockS3Client struct {
	buckets     map[string]map[string]mockObject
	pageSize    int
	deleteCalls int
}

func newMockS3() *mockS3Client {
	return &mockS3Client{
		buckets:  make(map[string]map[string]mockObject),
		pageSize: 1000,
	}
}

func (m *mockS3Client) ListBuckets(_ context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	var out []types.Bucket
	now := time.Now()
	for name := range m.buckets {
		out = append(out, types.Bucket{
			Name:         aws.String(name),
			CreationDate: aws.Time(now),
		})
	}
	return &s3.ListBucketsOutput{Buckets: out}, nil
}

func (m *mockS3Client) CreateBucket(_ context.Context, params *s3.CreateBucketInput, _ ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	b := aws.ToString(params.Bucket)
	if _, exists := m.buckets[b]; exists {
		return nil, fmt.Errorf("bucket %q already exists", b)
	}
	m.buckets[b] = make(map[string]mockObject)
	return &s3.CreateBucketOutput{}, nil
}

func (m *mockS3Client) DeleteBucket(_ context.Context, params *s3.DeleteBucketInput, _ ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	b := aws.ToString(params.Bucket)
	objs, exists := m.buckets[b]
	if !exists {
		return nil, fmt.Errorf("bucket %q does not exist", b)
	}
	if len(objs) > 0 {
		return nil, errors.New("bucket not empty")
	}
	delete(m.buckets, b)
	return &s3.DeleteBucketOutput{}, nil
}

func (m *mockS3Client) HeadBucket(_ context.Context, params *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	if _, exists := m.buckets[aws.ToString(params.Bucket)]; !exists {
		return nil, errors.New("not found")
	}
	return &s3.HeadBucketOutput{}, nil
}

func (m *mockS3Client) ListObjectsV2(_ context.Context, params *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	b := aws.ToString(params.Bucket)
	objs, exists := m.buckets[b]
	if !exists {
		return nil, fmt.Errorf("no such bucket %q", b)
	}
	prefix := aws.ToString(params.Prefix)
	var keys []string
	for k := range objs {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	start := 0
	if params.ContinuationToken != nil {
		_, _ = fmt.Sscanf(*params.ContinuationToken, "%d", &start)
	}
	end := min(start+m.pageSize, len(keys))

	var contents []types.Object
	for _, k := range keys[start:end] {
		o := objs[k]
		contents = append(contents, types.Object{
			Key:          aws.String(k),
			LastModified: aws.Time(o.lastModified),
			Size:         aws.Int64(int64(len(o.body))),
		})
	}

	out := &s3.ListObjectsV2Output{
		Contents: contents,
	}
	if end < len(keys) {
		out.IsTruncated = aws.Bool(true)
		out.NextContinuationToken = aws.String(fmt.Sprintf("%d", end))
	}
	return out, nil
}

func (m *mockS3Client) HeadObject(_ context.Context, params *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	objs, exists := m.buckets[aws.ToString(params.Bucket)]
	if !exists {
		return nil, errors.New("no such bucket")
	}
	if _, ok := objs[aws.ToString(params.Key)]; !ok {
		return nil, errors.New("no such key")
	}
	return &s3.HeadObjectOutput{}, nil
}

func (m *mockS3Client) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	b := aws.ToString(params.Bucket)
	objs, exists := m.buckets[b]
	if !exists {
		return nil, fmt.Errorf("no such bucket %q", b)
	}
	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	objs[aws.ToString(params.Key)] = mockObject{
		body:         data,
		contentType:  aws.ToString(params.ContentType),
		lastModified: time.Now(),
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObjects(_ context.Context, params *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	m.deleteCalls++
	b := aws.ToString(params.Bucket)
	objs, exists := m.buckets[b]
	if !exists {
		return nil, fmt.Errorf("no such bucket %q", b)
	}
	var deleted []types.DeletedObject
	for _, id := range params.Delete.Objects {
		k := aws.ToString(id.Key)
		if _, ok := objs[k]; ok {
			delete(objs, k)
			deleted = append(deleted, types.DeletedObject{Key: aws.String(k)})
		}
	}
	return &s3.DeleteObjectsOutput{Deleted: deleted}, nil
}

func TestManagerUploadFilePreservesRealContents(t *testing.T) {
	ctx := context.Background()
	mock := newMockS3()
	mgr := gos3.New(mock, nil)

	tmp := t.TempDir()
	content := "exact file payload that must be uploaded to S3"
	filePath := filepath.Join(tmp, "sample.txt")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Upload with Create=true into non-existent bucket.
	_, err := mgr.UploadFile(ctx, "backups", filePath, "daily", gos3.UploadOptions{Create: true})
	if err != nil {
		t.Fatalf("UploadFile failed: %v", err)
	}

	obj, ok := mock.buckets["backups"]["daily/sample.txt"]
	if !ok {
		t.Fatalf("expected object daily/sample.txt in bucket, got keys: %v", mock.buckets["backups"])
	}
	if string(obj.body) != content {
		t.Fatalf("uploaded body = %q, want %q", string(obj.body), content)
	}
	if !strings.HasPrefix(obj.contentType, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", obj.contentType)
	}

	// Second upload without Replace must fail.
	if _, err := mgr.UploadFile(ctx, "backups", filePath, "daily", gos3.UploadOptions{Replace: false}); err == nil {
		t.Fatal("expected error when overwriting without Replace=true")
	}

	// Second upload with Replace=true must succeed.
	if _, err := mgr.UploadFile(ctx, "backups", filePath, "daily", gos3.UploadOptions{Replace: true}); err != nil {
		t.Fatalf("expected Replace=true to succeed, got: %v", err)
	}
}

func TestManagerUploadFilesWithZip(t *testing.T) {
	ctx := context.Background()
	mock := newMockS3()
	_, _ = mock.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String("backups")})
	mgr := gos3.New(mock, nil)

	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "one.txt")
	f2 := filepath.Join(tmp, "two.txt")
	_ = os.WriteFile(f1, []byte("first"), 0o600)
	_ = os.WriteFile(f2, []byte("second"), 0o600)

	err := mgr.UploadFiles(ctx, "backups", []string{f1, f2}, "archives", gos3.UploadOptions{
		ZipName: "bundle",
	})
	if err != nil {
		t.Fatalf("UploadFiles with zip failed: %v", err)
	}

	obj, ok := mock.buckets["backups"]["archives/bundle.zip"]
	if !ok {
		t.Fatalf("expected archives/bundle.zip in bucket, got: %v", mock.buckets["backups"])
	}

	zr, err := zip.NewReader(bytes.NewReader(obj.body), int64(len(obj.body)))
	if err != nil {
		t.Fatalf("uploaded payload is not a valid zip: %v", err)
	}
	if len(zr.File) != 2 {
		t.Fatalf("expected 2 entries in zip, got %d", len(zr.File))
	}
}

func TestManagerCleanOlderAndBatching(t *testing.T) {
	ctx := context.Background()
	mock := newMockS3()
	mock.pageSize = 2
	mock.buckets["backups"] = map[string]mockObject{
		"daily/old1.txt":   {body: []byte("a"), lastModified: time.Now().AddDate(0, 0, -5)},
		"daily/old2.txt":   {body: []byte("b"), lastModified: time.Now().AddDate(0, -2, 0)},
		"daily/recent.txt": {body: []byte("c"), lastModified: time.Now().Add(-1 * time.Hour)},
		"other/old3.txt":   {body: []byte("d"), lastModified: time.Now().AddDate(0, 0, -10)},
	}

	mgr := gos3.New(mock, nil)

	// Invalid duration must return an error.
	if _, err := mgr.CleanOlder(ctx, "backups", "invalid", "daily"); err == nil {
		t.Fatal("expected error for invalid duration")
	}

	// Clean files older than 2 days in "daily" folder.
	deleted, err := mgr.CleanOlder(ctx, "backups", "2d", "daily")
	if err != nil {
		t.Fatalf("CleanOlder failed: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("expected 2 deleted objects, got %d", len(deleted))
	}

	if _, ok := mock.buckets["backups"]["daily/recent.txt"]; !ok {
		t.Error("daily/recent.txt should have been preserved")
	}
	if _, ok := mock.buckets["backups"]["other/old3.txt"]; !ok {
		t.Error("other/old3.txt outside folder should have been preserved")
	}
}

func TestManagerRemoveBucketForceAndPagination(t *testing.T) {
	ctx := context.Background()
	mock := newMockS3()
	mock.pageSize = 2
	mock.buckets["b"] = map[string]mockObject{
		"k1": {body: []byte("1"), lastModified: time.Now()},
		"k2": {body: []byte("2"), lastModified: time.Now()},
		"k3": {body: []byte("3"), lastModified: time.Now()},
	}
	mgr := gos3.New(mock, nil)

	if _, err := mgr.RemoveBucket(ctx, "b", false); err == nil {
		t.Fatal("expected error removing non-empty bucket without force")
	}
	if _, err := mgr.RemoveBucket(ctx, "b", true); err != nil {
		t.Fatalf("RemoveBucket force=true failed: %v", err)
	}
	if mgr.BucketExists(ctx, "b") {
		t.Fatal("expected bucket to be deleted")
	}
}

func TestManagerUploadMySQLMissingEnv(t *testing.T) {
	mgr := gos3.New(newMockS3(), nil)
	err := mgr.UploadMySQL(context.Background(), "b", "mysql", config.MySQLConfig{}, gos3.UploadOptions{})
	if err == nil {
		t.Fatal("expected error when MySQL credentials are missing")
	}
}
