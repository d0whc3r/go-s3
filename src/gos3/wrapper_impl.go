package gos3

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	gos3Buckets "s3/src/bucket"
	"s3/src/config"
	gos3Files "s3/src/file"
	gos3Shared "s3/src/shared"
)

func New(options *S3Config) S3Wrapper {
	cfg := config.Config()

	var s3Wrapper S3Wrapper
	s3Wrapper.Bucket = cfg.Bucket
	if options.Bucket != nil {
		s3Wrapper.Bucket = *options.Bucket
	}

	awsConfig := getAwsConfig(options)
	s3Wrapper.Endpoint = *awsConfig.Endpoint
	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		log.Fatal(err)
	}
	s3Wrapper.s3 = s3.New(sess)
	s3Wrapper.s3Buckets = gos3Buckets.New(s3Wrapper.s3)
	s3Wrapper.s3Files = gos3Files.New(s3Wrapper.s3)
	s3Wrapper.s3SharedFiles = gos3Shared.NewFiles(s3Wrapper.s3)
	s3Wrapper.s3SharedBuckets = gos3Shared.NewBuckets(s3Wrapper.s3)
	return s3Wrapper
}

func (w S3Wrapper) CreateBucket(bucket *string) (*s3.CreateBucketOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3SharedBuckets.CreateBucket(bucketName)
}

func (w S3Wrapper) GetBuckets() ([]*s3.Bucket, error) {
	return w.s3Buckets.GetBuckets()
}

func (w S3Wrapper) BucketExist(bucket string) bool {
	return w.s3SharedBuckets.BucketExist(bucket)
}

func (w S3Wrapper) RemoveBucket(force bool, bucket *string) (*s3.DeleteBucketOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Buckets.RemoveBucket(bucketName, force)
}

func (w S3Wrapper) GetFiles(bucket *string) ([]*s3.Object, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3SharedFiles.GetFiles(bucketName)
}

func (w S3Wrapper) UploadFile(file string, folder string, options *gos3Files.UploadOptions, bucket *string) (*s3.PutObjectOutput, error) {
	bucketName := w.getBucketName(bucket)
	return w.s3Files.UploadFile(bucketName, file, folder, options)
}
