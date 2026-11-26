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

// Package dbdump provides pure-Go database backup dumpers for MySQL/MariaDB,
// PostgreSQL, SQLite, and MongoDB without requiring external CLI binaries.
package dbdump

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/d0whc3r/go-s3/internal/config"
)

// Engine identifies a supported database engine.
type Engine string

const (
	// EngineMySQL supports MySQL and MariaDB servers.
	EngineMySQL Engine = "mysql"
	// EnginePostgres supports PostgreSQL servers.
	EnginePostgres Engine = "postgres"
	// EngineSQLite supports local SQLite database files.
	EngineSQLite Engine = "sqlite"
	// EngineMongoDB supports MongoDB servers.
	EngineMongoDB Engine = "mongodb"
)

// Dumper creates a point-in-time backup file in outDir and returns its path.
type Dumper interface {
	Name() string
	Dump(ctx context.Context, outDir string, now time.Time) (string, error)
}

// ParseEngine normalizes a user-supplied database engine name or alias.
func ParseEngine(raw string) (Engine, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "mysql", "mariadb":
		return EngineMySQL, nil
	case "postgres", "postgresql", "pg":
		return EnginePostgres, nil
	case "sqlite", "sqlite3":
		return EngineSQLite, nil
	case "mongo", "mongodb":
		return EngineMongoDB, nil
	default:
		return "", fmt.Errorf("unsupported database engine %q (supported: mysql, mariadb, postgres, sqlite, mongodb)", raw)
	}
}

// NewDumper constructs the appropriate Dumper for engine using cfg.
func NewDumper(engine Engine, cfg config.DatabaseConfig, logger *slog.Logger) (Dumper, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	switch engine {
	case EngineMySQL:
		return &MySQLDumper{cfg: cfg.MySQL, logger: logger}, nil
	case EnginePostgres:
		return &PostgresDumper{cfg: cfg.Postgres, logger: logger}, nil
	case EngineSQLite:
		return &SQLiteDumper{cfg: cfg.SQLite, logger: logger}, nil
	case EngineMongoDB:
		return &MongoDumper{cfg: cfg.MongoDB, logger: logger}, nil
	default:
		return nil, fmt.Errorf("unsupported database engine %q", engine)
	}
}
