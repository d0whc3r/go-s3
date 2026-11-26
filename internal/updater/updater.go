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

// Package updater implements self-updating gos3 from GitHub Releases with
// SHA-256 verification and atomic binary replacement.
package updater

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const defaultBaseURL = "https://github.com/d0whc3r/go-s3"

// Result summarizes the outcome of a self-update attempt.
type Result struct {
	CurrentVersion string
	LatestVersion  string
	Updated        bool
	TargetPath     string
}

// Updater downloads and installs newer gos3 releases.
type Updater struct {
	BaseURL    string
	HTTPClient *http.Client
	GOOS       string
	GOARCH     string
	Logger     *slog.Logger
}

// New creates an Updater configured for the current OS/architecture.
func New(logger *slog.Logger) *Updater {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Updater{
		BaseURL: defaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
		Logger: logger,
	}
}

// Update checks for a newer release and atomically replaces targetPath if an
// update is available. If targetPath is empty, the running executable is used.
func (u *Updater) Update(ctx context.Context, currentVersion, targetPath string) (Result, error) {
	if u.Logger == nil {
		u.Logger = slog.New(slog.DiscardHandler)
	}
	current := strings.TrimPrefix(strings.TrimSpace(currentVersion), "v")
	res := Result{CurrentVersion: current}

	latest, err := u.resolveLatestVersion(ctx)
	if err != nil {
		return res, err
	}
	res.LatestVersion = latest

	if current != "" && current == latest {
		u.Logger.Debug("already on latest version", "version", current)
		return res, nil
	}

	if targetPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return res, fmt.Errorf("locate current executable: %w", err)
		}
		resolved, err := filepath.EvalSymlinks(exe)
		if err == nil {
			exe = resolved
		}
		targetPath = exe
	}
	res.TargetPath = targetPath

	asset := u.assetName(latest)
	downloadBase := fmt.Sprintf("%s/releases/download/v%s", strings.TrimRight(u.BaseURL, "/"), latest)

	u.Logger.Debug("fetching checksums", "url", downloadBase+"/SHA256SUMS.txt")
	expectedHash, err := u.fetchExpectedChecksum(ctx, downloadBase+"/SHA256SUMS.txt", asset)
	if err != nil {
		return res, err
	}

	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, ".gos3-update-*")
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return res, fmt.Errorf("permission denied writing to %s (try: sudo gos3 --update): %w", dir, err)
		}
		return res, fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	u.Logger.Debug("downloading release binary", "url", downloadBase+"/"+asset)
	actualHash, err := u.downloadAndHash(ctx, downloadBase+"/"+asset, tmpFile)
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return res, err
	}

	if expectedHash != "" && !strings.EqualFold(actualHash, expectedHash) {
		return res, fmt.Errorf("checksum mismatch for %s: expected %s, got %s", asset, expectedHash, actualHash)
	}
	u.Logger.Debug("verified SHA-256 checksum", "asset", asset, "sha256", actualHash)

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return res, fmt.Errorf("set executable permissions: %w", err)
	}

	if err := replaceBinary(tmpPath, targetPath, u.GOOS); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return res, fmt.Errorf("permission denied replacing %s (try: sudo gos3 --update): %w", targetPath, err)
		}
		return res, fmt.Errorf("replace %s: %w", targetPath, err)
	}

	cleanup = false
	res.Updated = true
	return res, nil
}

func (u *Updater) assetName(ver string) string {
	name := fmt.Sprintf("gos3_v%s_%s-%s", ver, u.GOOS, u.GOARCH)
	if u.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func (u *Updater) resolveLatestVersion(ctx context.Context) (string, error) {
	latestURL := strings.TrimRight(u.BaseURL, "/") + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("build latest release request: %w", err)
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("check latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %d checking latest release", resp.StatusCode)
	}

	finalURL := resp.Request.URL.String()
	const marker = "/tag/"
	idx := strings.LastIndex(finalURL, marker)
	if idx == -1 {
		return "", fmt.Errorf("could not parse latest version tag from %q", finalURL)
	}
	tag := strings.TrimSpace(finalURL[idx+len(marker):])
	tag = strings.TrimPrefix(tag, "v")
	if tag == "" {
		return "", fmt.Errorf("empty version tag in %q", finalURL)
	}
	return tag, nil
}

func (u *Updater) fetchExpectedChecksum(ctx context.Context, sumsURL, asset string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumsURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download SHA256SUMS.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status %d downloading SHA256SUMS.txt", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			name := strings.TrimPrefix(fields[1], "*")
			if name == asset {
				return fields[0], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read SHA256SUMS.txt: %w", err)
	}
	return "", fmt.Errorf("checksum for %q not found in SHA256SUMS.txt", asset)
}

func (u *Updater) downloadAndHash(ctx context.Context, url string, dst io.Writer) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s failed with HTTP %d", url, resp.StatusCode)
	}

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, h), resp.Body); err != nil {
		return "", fmt.Errorf("write binary: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func replaceBinary(tmpPath, targetPath, goos string) error {
	if goos == "windows" {
		backup := targetPath + ".old"
		_ = os.Remove(backup)
		if err := os.Rename(targetPath, backup); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if err := os.Rename(tmpPath, targetPath); err != nil {
			_ = os.Rename(backup, targetPath)
			return err
		}
		_ = os.Remove(backup)
		return nil
	}
	return os.Rename(tmpPath, targetPath)
}
