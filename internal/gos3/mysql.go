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
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"

	"github.com/d0whc3r/go-s3/internal/config"
)

// UploadMySQL dumps the configured MySQL database into a temporary SQL file
// and uploads it (optionally zipped) to bucket under folder.
func (m *Manager) UploadMySQL(ctx context.Context, bucket, folder string, mysqlCfg config.MySQLConfig, opts UploadOptions) error {
	if mysqlCfg.User == "" || mysqlCfg.Password == "" || mysqlCfg.Database == "" {
		return errors.New("mysql backup requires $MYSQL_USER, $MYSQL_PASSWORD, and $MYSQL_DATABASE environment variables")
	}

	tmpDir, err := os.MkdirTemp("", "gos3-mysql-*")
	if err != nil {
		return fmt.Errorf("create temp directory for mysql dump: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	m.logger.Debug("connecting to mysql",
		"host", mysqlCfg.Host,
		"port", mysqlCfg.Port,
		"user", mysqlCfg.User,
		"database", mysqlCfg.Database,
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		mysqlCfg.User,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.Port,
		mysqlCfg.Database,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql server %s:%d: %w", mysqlCfg.Host, mysqlCfg.Port, err)
	}

	dumpName := fmt.Sprintf("mysqldump-%s-%s", mysqlCfg.Database, m.now().Format("2006-01-02.150405"))
	dumper, err := mysqldump.Register(db, tmpDir, dumpName)
	if err != nil {
		return fmt.Errorf("register mysql dumper: %w", err)
	}
	defer dumper.Close()

	m.logger.Debug("creating mysql dump", "dir", tmpDir, "name", dumpName)
	dumpPath, err := dumper.Dump()
	if err != nil {
		return fmt.Errorf("dump mysql database %q: %w", mysqlCfg.Database, err)
	}

	return m.UploadFiles(ctx, bucket, []string{dumpPath}, folder, opts)
}
