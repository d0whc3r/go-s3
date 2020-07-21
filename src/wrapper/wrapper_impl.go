package wrapper

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"s3/src/config"
	"s3/src/gos3"
)

func New(options *config.S3Config) S3Wrapper {
	cfg := config.Config()

	var s3Wrapper S3Wrapper
	s3Wrapper.Bucket = cfg.Bucket
	if options.Bucket != nil {
		s3Wrapper.Bucket = *options.Bucket
	}

	awsConfig := config.AwsConfig(options)
	s3Wrapper.Endpoint = *awsConfig.Endpoint
	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		log.Fatal(err)
	}
	s3Wrapper.s3 = s3.New(sess)
	s3Wrapper.s3Manager = gos3.New(s3Wrapper.s3)
	return s3Wrapper
}

func (w S3Wrapper) CreateBucket(bucket *string) (*s3.CreateBucketOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Manager.CreateBucket(bucketName)
}

func (w S3Wrapper) GetBuckets() ([]*s3.Bucket, error) {
	return w.s3Manager.GetBuckets()
}

func (w S3Wrapper) BucketExist(bucket string) bool {
	return w.s3Manager.BucketExist(bucket)
}

func (w S3Wrapper) RemoveBucket(force bool, bucket *string) (*s3.DeleteBucketOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Manager.RemoveBucket(bucketName, force)
}

func (w S3Wrapper) GetFiles(bucket *string) ([]*s3.Object, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Manager.GetFiles(bucketName)
}

func (w S3Wrapper) UploadFile(file string, folder string, options *gos3.UploadOptions, bucket *string) (*s3.PutObjectOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Manager.UploadFile(bucketName, file, folder, options)
}