package config

func Config() Configuration {
	return Configuration{
		Endpoint:       getEnv("ENDPOINT", ""),
		Bucket:         getEnv("BUCKET", ""),
		Region:         getEnv("REGION", "fake"),
		AccessKey:      getEnv("ACCESS_KEY", ""),
		SecretKey:      getEnv("SECRET_KEY", ""),
		MaxRetries:     getEnvInt("MAX_RETRIES", "3"),
		ForcePathStyle: getEnvBool("FORCE_PATH_STYLE", "true"),
		SslEnabled:     getEnvBool("SSL_ENABLED", "false"),
		Tag:            getEnv("TAG", "[go-s3]"),
		MysqlHost:      getEnv("MYSQL_HOST", "localhost"),
		MysqlPort:      getEnvInt("MYSQL_PORT", "3306"),
		MysqlUser:      getEnv("MYSQL_USER", ""),
		MysqlPassword:  getEnv("MYSQL_PASSWORD", ""),
		MysqlDatabase:  getEnv("MYSQL_DATABASE", ""),
	}
}
