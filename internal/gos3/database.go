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
	"context"
	"fmt"
	"os"

	"github.com/d0whc3r/go-s3/internal/dbdump"
)

// UploadDatabase runs dumper into a temporary directory and uploads the
// resulting dump file (optionally zipped) to bucket under folder.
func (m *Manager) UploadDatabase(ctx context.Context, bucket, folder string, dumper dbdump.Dumper, opts UploadOptions) error {
	tmpDir, err := os.MkdirTemp("", "gos3-db-*")
	if err != nil {
		return fmt.Errorf("create temp directory for %s dump: %w", dumper.Name(), err)
	}
	defer os.RemoveAll(tmpDir)

	dumpPath, err := dumper.Dump(ctx, tmpDir, m.now())
	if err != nil {
		return err
	}
	return m.UploadFiles(ctx, bucket, []string{dumpPath}, folder, opts)
}
