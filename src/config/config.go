package config

import (
	"os"
	"strconv"
)

type Configuration struct {
	Endpoint       string
	Bucket         string
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

func Config() Configuration {
	return Configuration{
		Endpoint:       getEnv("ENDPOINT",""),
		Bucket:         getEnv("BUCKET",""),
		AccessKey:      getEnv("ACCESS_KEY",""),
		SecretKey:      getEnv("SECRET_KEY",""),
		MaxRetries:     getEnvInt("MAX_RETRIES","3"),
		ForcePathStyle: getEnvBool("FORCE_PATH_STYLE","true"),
		SslEnabled:     getEnvBool("SSL_ENABLED","false"),
		Tag:            getEnv("TAG","[go-s3]"),
		MysqlHost:      getEnv("MYSQL_HOST","localhost"),
		MysqlPort:      getEnvInt("MYSQL_PORT","3306"),
		MysqlUser:      getEnv("MYSQL_USER",""),
		MysqlPassword:  getEnv("MYSQL_PASSWORD",""),
		MysqlDatabase:  getEnv("MYSQL_DATABASE",""),
	}
}
