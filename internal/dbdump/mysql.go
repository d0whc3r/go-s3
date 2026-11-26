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
	"time"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"

	"github.com/d0whc3r/go-s3/internal/config"
)

// MySQLDumper dumps MySQL and MariaDB databases to SQL files.
type MySQLDumper struct {
	cfg    config.MySQLConfig
	logger *slog.Logger
}

// Name returns the display name of the engine.
func (d *MySQLDumper) Name() string { return "MySQL" }

// Dump connects to MySQL/MariaDB and writes a SQL dump into outDir.
func (d *MySQLDumper) Dump(ctx context.Context, outDir string, now time.Time) (string, error) {
	if d.cfg.User == "" || d.cfg.Password == "" || d.cfg.Database == "" {
		return "", errors.New("mysql backup requires $MYSQL_USER, $MYSQL_PASSWORD, and $MYSQL_DATABASE environment variables")
	}

	d.logger.Debug("connecting to mysql",
		"host", d.cfg.Host,
		"port", d.cfg.Port,
		"user", d.cfg.User,
		"database", d.cfg.Database,
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		d.cfg.User,
		d.cfg.Password,
		d.cfg.Host,
		d.cfg.Port,
		d.cfg.Database,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", fmt.Errorf("open mysql connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return "", fmt.Errorf("ping mysql server %s:%d: %w", d.cfg.Host, d.cfg.Port, err)
	}

	dumpName := fmt.Sprintf("mysqldump-%s-%s", d.cfg.Database, now.Format("2006-01-02.150405"))
	dumper, err := mysqldump.Register(db, outDir, dumpName)
	if err != nil {
		return "", fmt.Errorf("register mysql dumper: %w", err)
	}
	defer dumper.Close()

	d.logger.Debug("creating mysql dump", "dir", outDir, "name", dumpName)
	dumpPath, err := dumper.Dump()
	if err != nil {
		return "", fmt.Errorf("dump mysql database %q: %w", d.cfg.Database, err)
	}
	return dumpPath, nil
}
