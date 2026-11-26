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

//go:build integration

package tests

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/dbdump"
	"github.com/d0whc3r/go-s3/internal/gos3"
)

func randomBucketName(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return "s3-test-" + hex.EncodeToString(b)
}

func TestIntegrationS3AndDatabases(t *testing.T) {
	_ = godotenv.Load("../test.env")

	ctx := context.Background()
	bucket := randomBucketName(t)
	client, cfg, err := config.NewS3Client(ctx, &config.Options{Bucket: &bucket}, nil)
	if err != nil {
		t.Fatalf("init S3 client: %v", err)
	}

	mgr := gos3.New(client, nil)
	if _, err := mgr.CreateBucket(ctx, bucket); err != nil {
		t.Fatalf("create bucket: %v", err)
	}
	defer func() {
		_, _ = mgr.RemoveBucket(ctx, bucket, true)
	}()

	if err := mgr.UploadFiles(ctx, bucket, []string{"./sample/sample1.txt", "./sample/sample2.jpg"}, "integration", gos3.UploadOptions{}); err != nil {
		t.Fatalf("upload files: %v", err)
	}

	files, err := mgr.ListFiles(ctx, bucket, "integration/")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	time.Sleep(2 * time.Second)
	deleted, err := mgr.CleanOlder(ctx, bucket, "1s", "integration")
	if err != nil {
		t.Fatalf("clean older: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("expected 2 deleted files, got %d", len(deleted))
	}

	if cfg.Databases.MySQL.User != "" && cfg.Databases.MySQL.Database != "" {
		dumper, err := dbdump.NewDumper(dbdump.EngineMySQL, cfg.Databases, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := mgr.UploadDatabase(ctx, bucket, "mysql", dumper, gos3.UploadOptions{Compress: true}); err != nil {
			t.Fatalf("upload mysql dump: %v", err)
		}
	}

	if cfg.Databases.Postgres.URL != "" || (cfg.Databases.Postgres.User != "" && cfg.Databases.Postgres.Database != "") {
		dumper, err := dbdump.NewDumper(dbdump.EnginePostgres, cfg.Databases, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := mgr.UploadDatabase(ctx, bucket, "postgres", dumper, gos3.UploadOptions{Compress: true}); err != nil {
			t.Fatalf("upload postgres dump: %v", err)
		}
	}

	if cfg.Databases.MongoDB.Database != "" {
		dumper, err := dbdump.NewDumper(dbdump.EngineMongoDB, cfg.Databases, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := mgr.UploadDatabase(ctx, bucket, "mongodb", dumper, gos3.UploadOptions{Compress: true}); err != nil {
			t.Fatalf("upload mongodb dump: %v", err)
		}
	}
}
