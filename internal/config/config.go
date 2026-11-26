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

// Package config loads and resolves gos3 configuration from environment
// variables and CLI overrides, and constructs AWS SDK v2 S3 clients.
package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Configuration holds all runtime settings resolved from environment variables
// and optional CLI flag overrides.
type Configuration struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	MaxRetries     int
	ForcePathStyle bool
	SSLEnabled     bool
	Debug          bool
	Databases      DatabaseConfig
}

// DatabaseConfig groups connection settings for all supported database engines.
type DatabaseConfig struct {
	MySQL    MySQLConfig
	Postgres PostgresConfig
	SQLite   SQLiteConfig
	MongoDB  MongoDBConfig
}

// MySQLConfig holds connection parameters for MySQL / MariaDB backups.
type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// PostgresConfig holds connection parameters for PostgreSQL backups.
type PostgresConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Schema   string
}

// SQLiteConfig holds parameters for SQLite backups.
type SQLiteConfig struct {
	Path string
}

// MongoDBConfig holds connection parameters for MongoDB backups.
type MongoDBConfig struct {
	URI      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// Options contains optional overrides (typically from CLI flags) that take
// precedence over environment variables.
type Options struct {
	Endpoint       *string
	Bucket         *string
	Region         *string
	MaxRetries     *int
	ForcePathStyle *bool
	SSLEnabled     *bool
	Debug          *bool
	SQLitePath     *string
}

// Load reads the base configuration from environment variables with sensible defaults.
func Load() Configuration {
	return Configuration{
		Endpoint:       getEnv("ENDPOINT", ""),
		Bucket:         getEnv("BUCKET", ""),
		Region:         getEnv("REGION", "us-east-1"),
		AccessKey:      getEnv("ACCESS_KEY", ""),
		SecretKey:      getEnv("SECRET_KEY", ""),
		MaxRetries:     getEnvInt("MAX_RETRIES", 3),
		ForcePathStyle: getEnvBool("FORCE_PATH_STYLE", true),
		SSLEnabled:     getEnvBool("SSL_ENABLED", false),
		Debug:          getEnvBool("DEBUG", false),
		Databases: DatabaseConfig{
			MySQL: MySQLConfig{
				Host:     getEnv("MYSQL_HOST", "localhost"),
				Port:     getEnvInt("MYSQL_PORT", 3306),
				User:     getEnv("MYSQL_USER", ""),
				Password: getEnv("MYSQL_PASSWORD", ""),
				Database: getEnv("MYSQL_DATABASE", ""),
			},
			Postgres: PostgresConfig{
				URL:      getEnv("POSTGRES_URL", ""),
				Host:     getEnv("POSTGRES_HOST", "localhost"),
				Port:     getEnvInt("POSTGRES_PORT", 5432),
				User:     getEnv("POSTGRES_USER", ""),
				Password: getEnv("POSTGRES_PASSWORD", ""),
				Database: getEnv("POSTGRES_DATABASE", ""),
				SSLMode:  getEnv("POSTGRES_SSLMODE", "prefer"),
				Schema:   getEnv("POSTGRES_SCHEMA", "public"),
			},
			SQLite: SQLiteConfig{
				Path: getEnv("SQLITE_PATH", ""),
			},
			MongoDB: MongoDBConfig{
				URI:      getEnv("MONGO_URI", ""),
				Host:     getEnv("MONGO_HOST", "localhost"),
				Port:     getEnvInt("MONGO_PORT", 27017),
				User:     getEnv("MONGO_USER", ""),
				Password: getEnv("MONGO_PASSWORD", ""),
				Database: getEnv("MONGO_DATABASE", ""),
			},
		},
	}
}

// Resolve merges environment configuration with any non-nil CLI overrides.
func Resolve(opts *Options) Configuration {
	cfg := Load()
	if opts == nil {
		return cfg
	}
	if opts.Endpoint != nil && *opts.Endpoint != "" {
		cfg.Endpoint = *opts.Endpoint
	}
	if opts.Bucket != nil && *opts.Bucket != "" {
		cfg.Bucket = *opts.Bucket
	}
	if opts.Region != nil && *opts.Region != "" {
		cfg.Region = *opts.Region
	}
	if opts.MaxRetries != nil {
		cfg.MaxRetries = *opts.MaxRetries
	}
	if opts.ForcePathStyle != nil {
		cfg.ForcePathStyle = *opts.ForcePathStyle
	}
	if opts.SSLEnabled != nil {
		cfg.SSLEnabled = *opts.SSLEnabled
	}
	if opts.Debug != nil && *opts.Debug {
		cfg.Debug = true
	}
	if opts.SQLitePath != nil && *opts.SQLitePath != "" {
		cfg.Databases.SQLite.Path = *opts.SQLitePath
	}
	return cfg
}

// NormalizedEndpoint ensures the endpoint includes an HTTP/HTTPS scheme based
// on SSLEnabled when no scheme was explicitly provided.
func (c Configuration) NormalizedEndpoint() string {
	ep := strings.TrimSpace(c.Endpoint)
	if ep == "" {
		return ""
	}
	if strings.HasPrefix(ep, "http://") || strings.HasPrefix(ep, "https://") {
		return ep
	}
	if c.SSLEnabled {
		return "https://" + ep
	}
	return "http://" + ep
}

// NewS3Client validates the resolved configuration and creates an AWS SDK v2 S3 client.
func NewS3Client(ctx context.Context, opts *Options, logger *slog.Logger) (*s3.Client, Configuration, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	cfg := Resolve(opts)
	if cfg.Bucket == "" {
		return nil, cfg, errors.New("no bucket configured: set --bucket or $BUCKET")
	}

	endpoint := cfg.NormalizedEndpoint()
	if endpoint == "" {
		return nil, cfg, errors.New("no endpoint configured: set --endpoint or $ENDPOINT")
	}
	cfg.Endpoint = endpoint

	logger.Debug("resolved S3 configuration",
		"endpoint", cfg.Endpoint,
		"bucket", cfg.Bucket,
		"region", cfg.Region,
		"max_retries", cfg.MaxRetries,
		"force_path_style", cfg.ForcePathStyle,
		"ssl_enabled", cfg.SSLEnabled,
		"has_static_credentials", cfg.AccessKey != "" && cfg.SecretKey != "",
	)

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithRetryMaxAttempts(cfg.MaxRetries),
	}
	if cfg.AccessKey != "" || cfg.SecretKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, cfg, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.ForcePathStyle
	})

	return client, cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func getEnvBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}
