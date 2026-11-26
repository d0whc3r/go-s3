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

package gos3

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CreateZip creates a zip archive at destPath containing all files and
// directories matched by sources (supporting glob patterns).
func CreateZip(sources []string, destPath string) error {
	if len(sources) == 0 {
		return errors.New("no source files specified for zip archive")
	}

	expanded, err := expandSources(sources)
	if err != nil {
		return err
	}
	if len(expanded) == 0 {
		return errors.New("no files matched the provided paths")
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create zip file %q: %w", destPath, err)
	}

	zw := zip.NewWriter(outFile)
	added := make(map[string]struct{})

	for _, src := range expanded {
		if err := addPathToZip(zw, src, added); err != nil {
			_ = zw.Close()
			_ = outFile.Close()
			return err
		}
	}

	if err := zw.Close(); err != nil {
		_ = outFile.Close()
		return fmt.Errorf("finalize zip archive %q: %w", destPath, err)
	}
	if err := outFile.Close(); err != nil {
		return fmt.Errorf("close zip file %q: %w", destPath, err)
	}
	return nil
}

func expandSources(sources []string) ([]string, error) {
	var result []string
	for _, s := range sources {
		if strings.ContainsAny(s, "*?[") {
			matches, err := filepath.Glob(s)
			if err != nil {
				return nil, fmt.Errorf("invalid glob pattern %q: %w", s, err)
			}
			result = append(result, matches...)
		} else {
			result = append(result, s)
		}
	}
	return result, nil
}

func addPathToZip(zw *zip.Writer, src string, added map[string]struct{}) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %q: %w", src, err)
	}

	if !info.IsDir() {
		return addFileToZip(zw, src, filepath.Base(src), info, added)
	}

	baseDir := filepath.Dir(filepath.Clean(src))
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(baseDir, p)
		if err != nil {
			return err
		}
		return addFileToZip(zw, p, filepath.ToSlash(rel), fi, added)
	})
}

func addFileToZip(zw *zip.Writer, filePath, archiveName string, info fs.FileInfo, added map[string]struct{}) error {
	cleanName := strings.TrimPrefix(filepath.ToSlash(archiveName), "/")
	if _, exists := added[cleanName]; exists {
		return nil
	}
	added[cleanName] = struct{}{}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("create zip header for %q: %w", filePath, err)
	}
	header.Name = cleanName
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("write zip header for %q: %w", filePath, err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open %q: %w", filePath, err)
	}
	defer f.Close()

	if _, err := io.Copy(writer, f); err != nil {
		return fmt.Errorf("write %q to zip: %w", filePath, err)
	}
	return nil
}
