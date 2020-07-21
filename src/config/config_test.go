package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	fmt.Println("[BEFORE ALL] Tests")
	exit := m.Run()
	defer mainTearDown(exit)
}

func mainTearDown(exit int) {
	fmt.Println("[AFTER ALL] Tests")
	os.Exit(exit)
}

func TestConfigDefaultValues(t *testing.T) {
	config := Config()
	assert.Equal(t, "fake", config.Region)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, true, config.ForcePathStyle)
	assert.Equal(t, false, config.SslEnabled)
}

func TestConfigGetEnvNotExist(t *testing.T) {
	v := getEnv("NOT_DEFINED", "unknown")
	assert.Equal(t, "unknown", v)
}

func TestConfigGetEnv(t *testing.T) {
	_ = os.Setenv("DEFINED", "defined-value")
	v := getEnv("DEFINED", "fail")
	assert.Equal(t, "defined-value", v)
}

func TestConfigGetEnvIntNotExistWrongFallback(t *testing.T) {
	v := getEnvInt("NOT_DEFINED", "unknown")
	assert.Equal(t, 0, v)
}

func TestConfigGetEnvIntNotExist(t *testing.T) {
	v := getEnvInt("NOT_DEFINED", "99")
	assert.Equal(t, 99, v)
}

func TestConfigGetEnvInt(t *testing.T) {
	_ = os.Setenv("DEFINED", "500")
	v := getEnvInt("DEFINED", "99")
	assert.Equal(t, 500, v)
}

func TestConfigGetEnvBoolNotExistWrongFallback(t *testing.T) {
	v := getEnvBool("NOT_DEFINED", "unknown")
	assert.Equal(t, false, v)
}

func TestConfigGetEnvBoolNotExist(t *testing.T) {
	v := getEnvBool("NOT_DEFINED", "true")
	assert.Equal(t, true, v)
}

func TestConfigGetEnvBool(t *testing.T) {
	_ = os.Setenv("DEFINED", "0")
	v := getEnvBool("DEFINED", "TRUE")
	assert.Equal(t, false, v)
}

func TestAwsConfig(t *testing.T) {
	endpoint := "none"
	region := "the-region"
	maxRetries := 5
	forcePathStyle := false
	sslEnabled := true
	conf := AwsConfig(&S3Config{
		Endpoint:       &endpoint,
		Region:         &region,
		MaxRetries:     &maxRetries,
		ForcePathStyle: &forcePathStyle,
		SslEnabled:     &sslEnabled,
	})
	assert.Equal(t, endpoint, *conf.Endpoint)
	assert.Equal(t, region, *conf.Region)
	assert.Equal(t, maxRetries, *conf.MaxRetries)
	assert.Equal(t, forcePathStyle, *conf.S3ForcePathStyle)
	assert.Equal(t, !sslEnabled, *conf.DisableSSL)
}

func TestAwsConfigUsingConfigEnv(t *testing.T) {
	_ = os.Setenv("ENDPOINT", "none")
	_ = os.Setenv("REGION", "theregion")
	_ = os.Setenv("MAX_RETRIES", "5")
	_ = os.Setenv("FORCE_PATH_STYLE", "false")
	_ = os.Setenv("SSL_ENABLED", "true")
	conf := AwsConfig(&S3Config{})
	assert.Equal(t, "none", *conf.Endpoint)
	assert.Equal(t, "theregion", *conf.Region)
	assert.Equal(t, 5, *conf.MaxRetries)
	assert.Equal(t, false, *conf.S3ForcePathStyle)
	assert.Equal(t, false, *conf.DisableSSL)
}
