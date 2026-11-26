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

package dbdump_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/d0whc3r/go-s3/internal/config"
	"github.com/d0whc3r/go-s3/internal/dbdump"
)

func TestParseEngine(t *testing.T) {
	tests := []struct {
		input   string
		want    dbdump.Engine
		wantErr bool
	}{
		{"mysql", dbdump.EngineMySQL, false},
		{"MariaDB", dbdump.EngineMySQL, false},
		{"postgres", dbdump.EnginePostgres, false},
		{"PostgreSQL", dbdump.EnginePostgres, false},
		{"pg", dbdump.EnginePostgres, false},
		{"sqlite", dbdump.EngineSQLite, false},
		{"sqlite3", dbdump.EngineSQLite, false},
		{"mongo", dbdump.EngineMongoDB, false},
		{"mongodb", dbdump.EngineMongoDB, false},
		{"oracle", "", true},
		{"", "", true},
	}

	for _, tc := range tests {
		got, err := dbdump.ParseEngine(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseEngine(%q) expected error, got %q", tc.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseEngine(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseEngine(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSQLiteDumperEndToEnd(t *testing.T) {
	tmp := t.TempDir()
	srcDB := filepath.Join(tmp, "app.db")

	db, err := sql.Open("sqlite", srcDB)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
		INSERT INTO items (name) VALUES ('alpha'), ('beta');
	`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()

	outDir := filepath.Join(tmp, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}

	dumper, err := dbdump.NewDumper(dbdump.EngineSQLite, config.DatabaseConfig{
		SQLite: config.SQLiteConfig{Path: srcDB},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.November, 26, 10, 0, 0, 0, time.UTC)
	dumpPath, err := dumper.Dump(context.Background(), outDir, now)
	if err != nil {
		t.Fatalf("SQLite Dump failed: %v", err)
	}
	if !strings.HasSuffix(dumpPath, "sqlite-app-2026-11-26.100000.sqlite") {
		t.Errorf("unexpected dump filename: %q", dumpPath)
	}

	// Verify the snapshot is a valid SQLite DB with all rows intact.
	snap, err := sql.Open("sqlite", dumpPath)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Close()

	var count int
	if err := snap.QueryRow("SELECT COUNT(*) FROM items").Scan(&count); err != nil {
		t.Fatalf("query snapshot: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows in SQLite snapshot, got %d", count)
	}
}

func TestDumpersValidationErrors(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	now := time.Now()

	engines := []dbdump.Engine{
		dbdump.EngineMySQL,
		dbdump.EnginePostgres,
		dbdump.EngineSQLite,
		dbdump.EngineMongoDB,
	}

	for _, eng := range engines {
		dumper, err := dbdump.NewDumper(eng, config.DatabaseConfig{}, nil)
		if err != nil {
			t.Fatalf("NewDumper(%s) unexpected error: %v", eng, err)
		}
		if _, err := dumper.Dump(ctx, tmp, now); err == nil {
			t.Errorf("expected validation error for empty %s config", eng)
		}
	}
}
