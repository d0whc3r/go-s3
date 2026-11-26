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

package dbdump

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/d0whc3r/go-s3/internal/config"
)

// SQLiteDumper creates consistent point-in-time SQLite snapshots using VACUUM INTO.
type SQLiteDumper struct {
	cfg    config.SQLiteConfig
	logger *slog.Logger
}

// Name returns the display name of the engine.
func (d *SQLiteDumper) Name() string { return "SQLite" }

// Dump creates a consistent SQLite backup file in outDir via VACUUM INTO.
func (d *SQLiteDumper) Dump(ctx context.Context, outDir string, now time.Time) (string, error) {
	srcPath := strings.TrimSpace(d.cfg.Path)
	if srcPath == "" {
		return "", errors.New("sqlite backup requires --sqlite <path> or $SQLITE_PATH environment variable")
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("stat sqlite database %q: %w", srcPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("sqlite path %q is a directory, expected a database file", srcPath)
	}

	base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	destPath := filepath.Join(outDir, fmt.Sprintf("sqlite-%s-%s.sqlite", base, now.Format("2006-01-02.150405")))

	d.logger.Debug("creating sqlite snapshot", "source", srcPath, "dest", destPath)

	db, err := sql.Open("sqlite", srcPath)
	if err != nil {
		return "", fmt.Errorf("open sqlite database %q: %w", srcPath, err)
	}
	defer db.Close()

	escapedDest := strings.ReplaceAll(destPath, "'", "''")
	if _, err := db.ExecContext(ctx, fmt.Sprintf("VACUUM INTO '%s'", escapedDest)); err != nil {
		return "", fmt.Errorf("vacuum sqlite database %q into backup: %w", srcPath, err)
	}

	return destPath, nil
}
