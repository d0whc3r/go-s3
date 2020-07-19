package gos3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"

	gos3Buckets "s3/src/bucket"
	"s3/src/config"
	gos3Files "s3/src/file"
	gos3Shared "s3/src/shared"
)

type S3Wrapper struct {
	Bucket          string
	Endpoint        string
	s3              *s3.S3
	s3Buckets       gos3Buckets.S3WrapperBuckets
	s3Files         gos3Files.S3WrapperFiles
	s3SharedFiles   gos3Shared.S3SharedFiles
	s3SharedBuckets gos3Shared.S3SharedBuckets
}

type S3Config struct {
	Bucket         *string
	Endpoint       *string
	Region         *string
	MaxRetries     *int
	ForcePathStyle *bool
	SslEnabled     *bool
}

func getAwsConfig(options *S3Config) aws.Config {
	cfg := config.Config()
	disableSsl := !cfg.SslEnabled
	if options.SslEnabled != nil {
		disableSsl = *options.SslEnabled
	}
	region := cfg.Region
	if options.Region != nil {
		region = *options.Region
	}
	endpoint := cfg.Endpoint
	if options.Endpoint != nil {
		endpoint = *options.Endpoint
	}
	maxRetries := cfg.MaxRetries
	if options.MaxRetries != nil {
		maxRetries = *options.MaxRetries
	}
	forcePathStyle := cfg.ForcePathStyle
	if options.ForcePathStyle != nil {
		forcePathStyle = *options.ForcePathStyle
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

func (w S3Wrapper) getBucketName(bucket *string) string {
	var bucketName string
	if bucket != nil {
		bucketName = *bucket
	} else {
		bucketName = w.Bucket
	}
	return bucketName
}
