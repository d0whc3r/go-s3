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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/d0whc3r/go-s3/internal/gos3"
)

func TestCreateZipFilesAndDirectory(t *testing.T) {
	tmp := t.TempDir()

	fileA := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(fileA, []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tmp, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fileB := filepath.Join(subDir, "b.txt")
	if err := os.WriteFile(fileB, []byte("nested content"), 0o600); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(tmp, "archive.zip")
	if err := gos3.CreateZip([]string{fileA, subDir}, dest); err != nil {
		t.Fatalf("CreateZip failed: %v", err)
	}

	zr, err := zip.OpenReader(dest)
	if err != nil {
		t.Fatalf("open created zip: %v", err)
	}
	defer zr.Close()

	got := make(map[string]string)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		got[f.Name] = string(data)
	}

	if got["a.txt"] != "hello world" {
		t.Errorf("expected a.txt content 'hello world', got %q", got["a.txt"])
	}
	if got["nested/b.txt"] != "nested content" {
		t.Errorf("expected nested/b.txt content 'nested content', got %q", got["nested/b.txt"])
	}
}

func TestCreateZipErrors(t *testing.T) {
	tmp := t.TempDir()
	dest := filepath.Join(tmp, "out.zip")

	if err := gos3.CreateZip(nil, dest); err == nil {
		t.Fatal("expected error for empty sources")
	}
	if err := gos3.CreateZip([]string{filepath.Join(tmp, "does-not-exist.txt")}, dest); err == nil {
		t.Fatal("expected error for missing file")
	}
}
