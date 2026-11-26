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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/d0whc3r/go-s3/internal/config"
)

// MongoDumper dumps all collections of a MongoDB database to Extended JSON.
type MongoDumper struct {
	cfg    config.MongoDBConfig
	logger *slog.Logger
}

// Name returns the display name of the engine.
func (d *MongoDumper) Name() string { return "MongoDB" }

// Dump connects to MongoDB and streams all collections as canonical Extended JSON into outDir.
func (d *MongoDumper) Dump(ctx context.Context, outDir string, now time.Time) (string, error) {
	if d.cfg.Database == "" {
		return "", errors.New("mongodb backup requires $MONGO_DATABASE environment variable")
	}

	uri := d.buildURI()
	d.logger.Debug("connecting to mongodb",
		"host", d.cfg.Host,
		"port", d.cfg.Port,
		"database", d.cfg.Database,
		"using_uri", d.cfg.URI != "",
	)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return "", fmt.Errorf("connect to mongodb: %w", err)
	}
	defer func() {
		_ = client.Disconnect(context.Background())
	}()

	if err := client.Ping(ctx, nil); err != nil {
		return "", fmt.Errorf("ping mongodb server: %w", err)
	}

	db := client.Database(d.cfg.Database)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return "", fmt.Errorf("list mongodb collections in %q: %w", d.cfg.Database, err)
	}
	sort.Strings(collections)

	filename := fmt.Sprintf("mongodump-%s-%s.json", d.cfg.Database, now.Format("2006-01-02.150405"))
	destPath := filepath.Join(outDir, filename)

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create mongodb dump file: %w", err)
	}
	bw := bufio.NewWriter(f)

	if err := d.writeCollections(ctx, db, collections, bw); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("flush mongodb dump file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close mongodb dump file: %w", err)
	}

	return destPath, nil
}

func (d *MongoDumper) buildURI() string {
	if d.cfg.URI != "" {
		return d.cfg.URI
	}
	u := &url.URL{
		Scheme: "mongodb",
		Host:   fmt.Sprintf("%s:%d", d.cfg.Host, d.cfg.Port),
	}
	if d.cfg.User != "" {
		u.User = url.UserPassword(d.cfg.User, d.cfg.Password)
	}
	return u.String()
}

func (d *MongoDumper) writeCollections(ctx context.Context, db *mongo.Database, collections []string, w *bufio.Writer) error {
	if _, err := w.WriteString("{\n"); err != nil {
		return err
	}

	writtenCollections := 0
	for _, colName := range collections {
		if strings.HasPrefix(colName, "system.") {
			continue
		}
		if writtenCollections > 0 {
			if _, err := w.WriteString(",\n"); err != nil {
				return err
			}
		}
		writtenCollections++

		nameJSON, err := json.Marshal(colName)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  %s: [", nameJSON); err != nil {
			return err
		}

		cursor, err := db.Collection(colName).Find(ctx, bson.D{})
		if err != nil {
			return fmt.Errorf("query collection %q: %w", colName, err)
		}

		docCount := 0
		for cursor.Next(ctx) {
			if docCount > 0 {
				if _, err := w.WriteString(","); err != nil {
					_ = cursor.Close(ctx)
					return err
				}
			}
			extJSON, err := bson.MarshalExtJSON(cursor.Current, true, false)
			if err != nil {
				_ = cursor.Close(ctx)
				return fmt.Errorf("marshal document in %q: %w", colName, err)
			}
			if _, err := w.WriteString("\n    "); err != nil {
				_ = cursor.Close(ctx)
				return err
			}
			if _, err := w.Write(extJSON); err != nil {
				_ = cursor.Close(ctx)
				return err
			}
			docCount++
		}
		if err := cursor.Err(); err != nil {
			_ = cursor.Close(ctx)
			return fmt.Errorf("iterate collection %q: %w", colName, err)
		}
		_ = cursor.Close(ctx)

		if docCount > 0 {
			if _, err := w.WriteString("\n  ]"); err != nil {
				return err
			}
		} else {
			if _, err := w.WriteString("]"); err != nil {
				return err
			}
		}

		d.logger.Debug("dumped mongodb collection", "collection", colName, "documents", docCount)
	}

	_, err := w.WriteString("\n}\n")
	return err
}
