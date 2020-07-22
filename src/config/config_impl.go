package config

import (
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/credentials"
)

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
    MysqlHost:      getEnv("MYSQL_HOST", "localhost"),
    MysqlPort:      getEnvInt("MYSQL_PORT", "3306"),
    MysqlUser:      getEnv("MYSQL_USER", ""),
    MysqlPassword:  getEnv("MYSQL_PASSWORD", ""),
    MysqlDatabase:  getEnv("MYSQL_DATABASE", ""),
  }
}

func AwsConfig(options *S3Config) aws.Config {
  cfg := Config()
  disableSsl := !cfg.SslEnabled
  region := cfg.Region
  endpoint := cfg.Endpoint
  maxRetries := cfg.MaxRetries
  forcePathStyle := cfg.ForcePathStyle
  if options != nil {
    if options.SslEnabled != nil {
      disableSsl = !*options.SslEnabled
    }
    if options.Region != nil {
      region = *options.Region
    }
    if options.Endpoint != nil {
      endpoint = *options.Endpoint
    }
    if options.MaxRetries != nil {
      maxRetries = *options.MaxRetries
    }
    if options.ForcePathStyle != nil {
      forcePathStyle = *options.ForcePathStyle
    }
  }
  return aws.Config{
    CredentialsChainVerboseErrors:     nil,
    Credentials:                       credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
    Endpoint:                          &endpoint,
    EndpointResolver:                  nil,
    EnforceShouldRetryCheck:           nil,
    Region:                            &region,
    DisableSSL:                        &disableSsl,
    HTTPClient:                        nil,
    LogLevel:                          nil,
    Logger:                            nil,
    MaxRetries:                        &maxRetries,
    Retryer:                           nil,
    DisableParamValidation:            nil,
    DisableComputeChecksums:           nil,
    S3ForcePathStyle:                  &forcePathStyle,
    S3Disable100Continue:              nil,
    S3UseAccelerate:                   nil,
    S3DisableContentMD5Validation:     nil,
    S3UseARNRegion:                    nil,
    LowerCaseHeaderMaps:               nil,
    EC2MetadataDisableTimeoutOverride: nil,
    UseDualStack:                      nil,
    SleepDelay:                        nil,
    DisableRestProtocolURICleaning:    nil,
    EnableEndpointDiscovery:           nil,
    DisableEndpointHostPrefix:         nil,
    STSRegionalEndpoint:               0,
    S3UsEast1RegionalEndpoint:         0,
  }
}
