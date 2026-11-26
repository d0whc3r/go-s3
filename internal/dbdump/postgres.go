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
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/d0whc3r/go-s3/internal/config"
)

// PostgresDumper dumps PostgreSQL schemas and table data to a pure-Go SQL file.
type PostgresDumper struct {
	cfg    config.PostgresConfig
	logger *slog.Logger
}

// Name returns the display name of the engine.
func (d *PostgresDumper) Name() string { return "PostgreSQL" }

// Dump connects to PostgreSQL and writes a self-contained SQL dump into outDir.
func (d *PostgresDumper) Dump(ctx context.Context, outDir string, now time.Time) (string, error) {
	connStr, dbLabel, err := d.buildConnString()
	if err != nil {
		return "", err
	}

	schema := d.cfg.Schema
	if schema == "" {
		schema = "public"
	}

	d.logger.Debug("connecting to postgres",
		"host", d.cfg.Host,
		"port", d.cfg.Port,
		"database", dbLabel,
		"schema", schema,
		"using_url", d.cfg.URL != "",
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return "", fmt.Errorf("open postgres connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return "", fmt.Errorf("ping postgres server: %w", err)
	}

	filename := fmt.Sprintf("pgdump-%s-%s.sql", dbLabel, now.Format("2006-01-02.150405"))
	destPath := filepath.Join(outDir, filename)

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create postgres dump file: %w", err)
	}
	bw := bufio.NewWriter(f)

	if err := d.writeDump(ctx, db, bw, dbLabel, schema, now); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("flush postgres dump file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close postgres dump file: %w", err)
	}

	return destPath, nil
}

func (d *PostgresDumper) buildConnString() (string, string, error) {
	if d.cfg.URL != "" {
		u, err := url.Parse(d.cfg.URL)
		if err != nil {
			return "", "", fmt.Errorf("parse $POSTGRES_URL: %w", err)
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		if dbName == "" {
			dbName = "postgres"
		}
		return d.cfg.URL, dbName, nil
	}

	if d.cfg.User == "" || d.cfg.Database == "" {
		return "", "", errors.New("postgres backup requires $POSTGRES_URL or $POSTGRES_USER and $POSTGRES_DATABASE environment variables")
	}

	sslMode := d.cfg.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.cfg.User, d.cfg.Password),
		Host:   fmt.Sprintf("%s:%d", d.cfg.Host, d.cfg.Port),
		Path:   d.cfg.Database,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String(), d.cfg.Database, nil
}

func (d *PostgresDumper) writeDump(ctx context.Context, db *sql.DB, w *bufio.Writer, dbName, schema string, now time.Time) error {
	fmt.Fprintf(w, "-- gos3 PostgreSQL dump\n")
	fmt.Fprintf(w, "-- Database: %s | Schema: %s | Generated: %s\n\n", dbName, schema, now.UTC().Format(time.RFC3339))
	fmt.Fprintf(w, "BEGIN;\n\n")

	tables, err := pgListTables(ctx, db, schema)
	if err != nil {
		return err
	}
	d.logger.Debug("discovered postgres tables", "schema", schema, "count", len(tables))

	for _, table := range tables {
		if err := d.dumpTable(ctx, db, w, schema, table); err != nil {
			return fmt.Errorf("dump table %s.%s: %w", schema, table, err)
		}
	}

	fmt.Fprintf(w, "COMMIT;\n")
	return nil
}

func pgListTables(ctx context.Context, db *sql.DB, schema string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name`, schema)
	if err != nil {
		return nil, fmt.Errorf("list postgres tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (d *PostgresDumper) dumpTable(ctx context.Context, db *sql.DB, w *bufio.Writer, schema, table string) error {
	qualified := quotePGIdent(schema) + "." + quotePGIdent(table)

	colRows, err := db.QueryContext(ctx, `
		SELECT
			a.attname,
			pg_catalog.format_type(a.atttypid, a.atttypmod),
			a.attnotnull,
			COALESCE(pg_catalog.pg_get_expr(ad.adbin, ad.adrelid), '')
		FROM pg_catalog.pg_attribute a
		JOIN pg_catalog.pg_class c ON a.attrelid = c.oid
		JOIN pg_catalog.pg_namespace n ON c.relnamespace = n.oid
		LEFT JOIN pg_catalog.pg_attrdef ad ON a.attrelid = ad.adrelid AND a.attnum = ad.adnum
		WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`, schema, table)
	if err != nil {
		return err
	}
	defer colRows.Close()

	var colDefs []string
	var colNames []string
	for colRows.Next() {
		var name, dataType, defaultExpr string
		var notNull bool
		if err := colRows.Scan(&name, &dataType, &notNull, &defaultExpr); err != nil {
			return err
		}
		colNames = append(colNames, quotePGIdent(name))
		def := fmt.Sprintf("  %s %s", quotePGIdent(name), dataType)
		if defaultExpr != "" && !strings.HasPrefix(defaultExpr, "nextval(") {
			def += " DEFAULT " + defaultExpr
		}
		if notNull {
			def += " NOT NULL"
		}
		colDefs = append(colDefs, def)
	}
	if err := colRows.Err(); err != nil {
		return err
	}

	pkCols, err := pgPrimaryKeyColumns(ctx, db, schema, table)
	if err != nil {
		return err
	}
	if len(pkCols) > 0 {
		quotedPK := make([]string, len(pkCols))
		for i, c := range pkCols {
			quotedPK[i] = quotePGIdent(c)
		}
		colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(quotedPK, ", ")))
	}

	fmt.Fprintf(w, "-- Table: %s\n", qualified)
	fmt.Fprintf(w, "DROP TABLE IF EXISTS %s CASCADE;\n", qualified)
	fmt.Fprintf(w, "CREATE TABLE %s (\n%s\n);\n\n", qualified, strings.Join(colDefs, ",\n"))

	// Dump table rows.
	dataRows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM %s", strings.Join(colNames, ", "), qualified))
	if err != nil {
		return err
	}
	defer dataRows.Close()

	vals := make([]any, len(colNames))
	valPtrs := make([]any, len(colNames))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}

	insertPrefix := fmt.Sprintf("INSERT INTO %s (%s) VALUES (", qualified, strings.Join(colNames, ", "))
	for dataRows.Next() {
		if err := dataRows.Scan(valPtrs...); err != nil {
			return err
		}
		formatted := make([]string, len(vals))
		for i, v := range vals {
			formatted[i] = formatPGValue(v)
		}
		fmt.Fprintf(w, "%s%s);\n", insertPrefix, strings.Join(formatted, ", "))
	}
	if err := dataRows.Err(); err != nil {
		return err
	}

	// Dump non-primary-key indexes.
	idxRows, err := db.QueryContext(ctx, `
		SELECT indexdef
		FROM pg_indexes
		WHERE schemaname = $1 AND tablename = $2
		  AND indexname NOT IN (
			SELECT conname FROM pg_constraint
			WHERE contype = 'p' AND conrelid = ($1 || '.' || $2)::regclass
		  )
		ORDER BY indexname`, schema, table)
	if err == nil {
		defer idxRows.Close()
		for idxRows.Next() {
			var idxDef string
			if err := idxRows.Scan(&idxDef); err == nil && idxDef != "" {
				fmt.Fprintf(w, "%s;\n", idxDef)
			}
		}
	}

	fmt.Fprintf(w, "\n")
	return nil
}

func pgPrimaryKeyColumns(ctx context.Context, db *sql.DB, schema, table string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE i.indisprimary AND n.nspname = $1 AND c.relname = $2
		ORDER BY array_position(i.indkey, a.attnum)`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func quotePGIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

func formatPGValue(v any) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "TRUE"
		}
		return "FALSE"
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int:
		return strconv.Itoa(val)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'g', -1, 32)
	case []byte:
		return `'\\x` + hex.EncodeToString(val) + `'`
	case time.Time:
		return "'" + val.UTC().Format(time.RFC3339Nano) + "'"
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	default:
		return "'" + strings.ReplaceAll(fmt.Sprint(val), "'", "''") + "'"
	}
}
