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

package updater_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/d0whc3r/go-s3/internal/updater"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newFakeReleaseClient(latestTag, assetName string, payload []byte, checksumOverride string) *http.Client {
	sum := sha256.Sum256(payload)
	hash := hex.EncodeToString(sum[:])
	if checksumOverride != "" {
		hash = checksumOverride
	}

	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/releases/latest":
				return &http.Response{
					StatusCode: http.StatusFound,
					Header:     http.Header{"Location": []string{"/releases/tag/" + latestTag}},
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Request:    req,
				}, nil
			case "/releases/tag/" + latestTag:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Request:    req,
				}, nil
			case "/releases/download/" + latestTag + "/SHA256SUMS.txt":
				body := fmt.Sprintf("%s  %s\n", hash, assetName)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request:    req,
				}, nil
			case "/releases/download/" + latestTag + "/" + assetName:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(payload)),
					Request:    req,
				}, nil
			default:
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
					Request:    req,
				}, nil
			}
		}),
	}
}

func TestUpdaterAlreadyUpToDate(t *testing.T) {
	u := &updater.Updater{
		BaseURL:    "https://example.test",
		HTTPClient: newFakeReleaseClient("v2.1.0", "gos3_v2.1.0_linux-amd64", []byte("bin"), ""),
		GOOS:       "linux",
		GOARCH:     "amd64",
	}

	res, err := u.Update(context.Background(), "2.1.0", filepath.Join(t.TempDir(), "gos3"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Updated {
		t.Fatal("expected Updated=false when already on latest version")
	}
	if res.LatestVersion != "2.1.0" {
		t.Fatalf("expected LatestVersion=2.1.0, got %q", res.LatestVersion)
	}
}

func TestUpdaterDownloadsVerifiesAndReplaces(t *testing.T) {
	newBinary := []byte("#!/bin/sh\necho v2.1.0\n")
	target := filepath.Join(t.TempDir(), "gos3")
	if err := os.WriteFile(target, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	u := &updater.Updater{
		BaseURL:    "https://example.test",
		HTTPClient: newFakeReleaseClient("v2.1.0", "gos3_v2.1.0_linux-amd64", newBinary, ""),
		GOOS:       "linux",
		GOARCH:     "amd64",
	}

	res, err := u.Update(context.Background(), "2.0.0", target)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.Updated {
		t.Fatal("expected Updated=true")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(newBinary) {
		t.Fatalf("replaced binary content mismatch: got %q, want %q", string(got), string(newBinary))
	}
}

func TestUpdaterRejectsBadChecksum(t *testing.T) {
	target := filepath.Join(t.TempDir(), "gos3")
	if err := os.WriteFile(target, []byte("original-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	u := &updater.Updater{
		BaseURL:    "https://example.test",
		HTTPClient: newFakeReleaseClient("v2.1.0", "gos3_v2.1.0_linux-amd64", []byte("corrupted"), "0000000000000000000000000000000000000000000000000000000000000000"),
		GOOS:       "linux",
		GOARCH:     "amd64",
	}

	if _, err := u.Update(context.Background(), "2.0.0", target); err == nil {
		t.Fatal("expected checksum mismatch error")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original-binary" {
		t.Fatalf("existing binary should remain untouched on checksum failure, got %q", string(got))
	}
}
