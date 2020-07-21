package config

import (
	"os"
	"strconv"
)

type S3Config struct {
	Bucket         *string
	Endpoint       *string
	Region         *string
	MaxRetries     *int
	ForcePathStyle *bool
	SslEnabled     *bool
}

type Configuration struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	MaxRetries     int
	ForcePathStyle bool
	SslEnabled     bool
	MysqlHost      string
	MysqlPort      int
	MysqlUser      string
	MysqlPassword  string
	MysqlDatabase  string
}

func getEnv(key, fallback string) (value string) {
	if value = os.Getenv(key); len(value) == 0 {
		value = fallback
	}
	return
}

func getEnvInt(key, fallback string) (value int) {
	str := getEnv(key, fallback)
	var err error
	if value, err = strconv.Atoi(str); err != nil {
		value = 0
	}
	return
}

func getEnvBool(key, fallback string) bool {
	str := getEnv(key, fallback)
	switch str {
	case "true", "TRUE", "1":
		return true
	default:
		return false
	}
}
