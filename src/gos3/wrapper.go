package gos3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"s3/src/config"
)

type S3Wrapper struct {
	Bucket    string
	Endpoint  string
	s3        s3.S3
	s3Buckets S3WrapperBuckets
	s3Files   S3WrapperFiles
}

func (w *S3Wrapper) New() S3Wrapper {
	sess := session.Must(session.NewSession())
	var cfg = config.Config()
	var disableSsl = !cfg.SslEnabled
	w.s3 = *s3.New(sess, &aws.Config{
		CredentialsChainVerboseErrors:     nil,
		Credentials:                       credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:                          &w.Endpoint,
		EndpointResolver:                  nil,
		EnforceShouldRetryCheck:           nil,
		Region:                            nil,
		DisableSSL:                        &disableSsl,
		HTTPClient:                        nil,
		LogLevel:                          nil,
		Logger:                            nil,
		MaxRetries:                        &cfg.MaxRetries,
		Retryer:                           nil,
		DisableParamValidation:            nil,
		DisableComputeChecksums:           nil,
		S3ForcePathStyle:                  &cfg.ForcePathStyle,
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
	})
}

func (w *S3Wrapper) GetBuckets() {
	return w.s3Buckets.GetBuckets()
}
