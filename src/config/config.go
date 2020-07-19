package config

import (
	"os"
	"strconv"
)

type Configuration struct {
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	MaxRetries     int
	ForcePathStyle bool
	SslEnabled     bool
	Tag            string
	MysqlHost      string
	MysqlPort      int
	MysqlUser      string
	MysqlPassword  string
	MysqlDatabase  string
}

func getEnv(key, fallback string) (value string) {
	value = os.Getenv(key)
	if len(value) == 0 {
		value = fallback
	}
	return
}

func getEnvInt(key, fallback string) (value int) {
	var str = getEnv(key, fallback)
	value, err := strconv.Atoi(str)
	if err != nil {
		value = 0
	}
	return
}

func getEnvBool(key, fallback string) bool {
	var str = getEnv(key, fallback)
	var trues = []string{
		"true",
		"TRUE",
		"1",
	}
	for _, value := range trues {
		if str == value {
			return true
		}
	}
	return false
}
