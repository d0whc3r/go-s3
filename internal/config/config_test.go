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

package config_test

import (
	"context"
	"testing"

	"github.com/d0whc3r/go-s3/internal/config"
)

func TestLoadAndResolve(t *testing.T) {
	t.Setenv("ENDPOINT", "s3.example.com:9000")
	t.Setenv("BUCKET", "my-bucket")
	t.Setenv("REGION", "eu-central-1")
	t.Setenv("ACCESS_KEY", "ak")
	t.Setenv("SECRET_KEY", "sk")
	t.Setenv("MAX_RETRIES", "5")
	t.Setenv("FORCE_PATH_STYLE", "false")
	t.Setenv("SSL_ENABLED", "true")
	t.Setenv("DEBUG", "true")
	t.Setenv("MYSQL_HOST", "db.local")
	t.Setenv("MYSQL_PORT", "3307")
	t.Setenv("MYSQL_USER", "root")
	t.Setenv("MYSQL_PASSWORD", "secret")
	t.Setenv("MYSQL_DATABASE", "app")

	cfg := config.Load()
	if cfg.Endpoint != "s3.example.com:9000" {
		t.Fatalf("unexpected endpoint: %q", cfg.Endpoint)
	}
	if cfg.Bucket != "my-bucket" || cfg.Region != "eu-central-1" {
		t.Fatalf("unexpected bucket/region: %+v", cfg)
	}
	if cfg.MaxRetries != 5 || cfg.ForcePathStyle || !cfg.SSLEnabled || !cfg.Debug {
		t.Fatalf("unexpected flags/numeric config: %+v", cfg)
	}
	if cfg.MySQL.Host != "db.local" || cfg.MySQL.Port != 3307 || cfg.MySQL.User != "root" || cfg.MySQL.Password != "secret" || cfg.MySQL.Database != "app" {
		t.Fatalf("unexpected mysql config: %+v", cfg.MySQL)
	}

	otherEndpoint := "https://override.example.com"
	otherBucket := "override-bucket"
	otherRegion := "us-west-2"
	retries := 9
	forcePath := true
	ssl := false
	resolved := config.Resolve(&config.Options{
		Endpoint:       &otherEndpoint,
		Bucket:         &otherBucket,
		Region:         &otherRegion,
		MaxRetries:     &retries,
		ForcePathStyle: &forcePath,
		SSLEnabled:     &ssl,
	})
	if resolved.Endpoint != otherEndpoint || resolved.Bucket != otherBucket || resolved.Region != otherRegion {
		t.Fatalf("overrides not applied: %+v", resolved)
	}
	if resolved.MaxRetries != 9 || !resolved.ForcePathStyle || resolved.SSLEnabled {
		t.Fatalf("numeric/bool overrides not applied: %+v", resolved)
	}
}

func TestNormalizedEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   string
		sslEnabled bool
		want       string
	}{
		{"empty", "", false, ""},
		{"explicit http", "http://localhost:9000", true, "http://localhost:9000"},
		{"explicit https", "https://s3.amazonaws.com", false, "https://s3.amazonaws.com"},
		{"bare host http", "localhost:9000", false, "http://localhost:9000"},
		{"bare host https", "s3.amazonaws.com", true, "https://s3.amazonaws.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := config.Configuration{Endpoint: tc.endpoint, SSLEnabled: tc.sslEnabled}
			if got := c.NormalizedEndpoint(); got != tc.want {
				t.Errorf("NormalizedEndpoint() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewS3ClientValidation(t *testing.T) {
	t.Setenv("ENDPOINT", "")
	t.Setenv("BUCKET", "")
	t.Setenv("ACCESS_KEY", "test")
	t.Setenv("SECRET_KEY", "test")

	if _, _, err := config.NewS3Client(context.Background(), nil, nil); err == nil {
		t.Fatal("expected error when bucket is missing")
	}

	bucket := "b"
	if _, _, err := config.NewS3Client(context.Background(), &config.Options{Bucket: &bucket}, nil); err == nil {
		t.Fatal("expected error when endpoint is missing")
	}

	endpoint := "localhost:9000"
	client, cfg, err := config.NewS3Client(context.Background(), &config.Options{Bucket: &bucket, Endpoint: &endpoint}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil S3 client")
	}
	if cfg.Endpoint != "http://localhost:9000" {
		t.Fatalf("expected normalized endpoint, got %q", cfg.Endpoint)
	}
}
